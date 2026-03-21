package dataBase

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// -------------------------- 常量定义 --------------------------
const (
	DefaultDSN = "postgres://postgres:gg@localhost:5432/postgres?sslmode=disable"

	//Nodes -------------------------- Nodes 表相关 SQL --------------------------

	SqlNCreate = `
        WITH ins AS (
            INSERT INTO nodes (content)
                VALUES ($1)
                ON CONFLICT (content) DO NOTHING
                RETURNING id, content, created_at)
        SELECT id, content, created_at
        FROM ins
        UNION ALL
        SELECT id, content, created_at
        FROM nodes
        WHERE content = $1
        LIMIT 1;
    `

	SqlNFind      = `SELECT id, content, created_at FROM nodes WHERE id = $1;`
	SqlNAll       = `SELECT id, content, created_at FROM nodes;`
	SqlNFindByVal = `SELECT id, content, created_at FROM nodes WHERE content = $1;`

	// SqlPCreate -------------------------- Projection 表相关 SQL --------------------------
	SqlPCreate = `WITH ins AS (
        INSERT INTO projection (relation_id, node_id, nodeType)
            VALUES ($1, $2, $3)
            ON CONFLICT (relation_id, node_id, nodetype) DO NOTHING
            RETURNING id, relation_id, node_id, nodeType, created_at)
             SELECT id, relation_id, node_id, nodeType, created_at
             FROM ins
             UNION ALL
             SELECT id, relation_id, node_id, nodeType, created_at
             FROM projection
             WHERE relation_id = $1
               AND node_id = $2
             LIMIT 1;`
	SqlPDelete         = `DELETE FROM projection WHERE id = $1 RETURNING id, relation_id, node_id, nodeType, created_at;`
	SqlPFindByNode     = `SELECT id, relation_id, node_id, nodeType, created_at FROM projection WHERE node_id = $1;`
	SqlPFindByRelation = `SELECT id, relation_id, node_id, nodeType, created_at FROM projection WHERE relation_id = $1;`

	// SqlRCreate -------------------------- Relations 表相关 SQL --------------------------
	SqlRCreate = `INSERT INTO relations DEFAULT VALUES RETURNING id, created_at;`

	// SqlICreate -------------------------- Identity 表相关 SQL --------------------------
	SqlICreate = `WITH ins AS (
        INSERT INTO identity (k, v)
            VALUES ($1, $2)
            ON CONFLICT (k, v) DO NOTHING
            RETURNING id, k, v, created_at)
             SELECT id, k, v, created_at
             FROM ins
             UNION ALL
             SELECT id, k, v, created_at
             FROM identity
             WHERE k = $1
               AND v = $2
             LIMIT 1;`
	SqlIFindByK  = `SELECT id, k, v, created_at FROM identity WHERE k = $1;`
	SqlIFindByKv = `SELECT id, k, v, created_at FROM identity WHERE k = $1 AND v = $2;`
	SqlIUpdate   = `
        UPDATE identity 
        SET v = $2, created_at = CURRENT_TIMESTAMP
        WHERE k = $1
        RETURNING id, k, v, created_at;
    `

	// SqlInit -------------------------- 初始化和重置 SQL --------------------------
	SqlInit = `
    CREATE TABLE IF NOT EXISTS nodes
    (
        id         SERIAL PRIMARY KEY,
        content    TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS relations
    (
        id         SERIAL PRIMARY KEY,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    DO
    $$
        BEGIN
            IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'node_type_enum') THEN
                CREATE TYPE node_type_enum AS ENUM ('trigger', 'result');
            END IF;
        END
    $$;
    CREATE TABLE IF NOT EXISTS projection
    (
        id          SERIAL PRIMARY KEY,
        relation_id INTEGER        NOT NULL,
        node_id     INTEGER        NOT NULL,
        nodeType    node_type_enum NOT NULL,
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE (relation_id, node_id, nodeType),
        FOREIGN KEY (relation_id) REFERENCES relations (id) ON DELETE CASCADE,
        FOREIGN KEY (node_id) REFERENCES nodes (id) ON DELETE CASCADE
    );
    CREATE TABLE IF NOT EXISTS identity
    (
        id         SERIAL PRIMARY KEY,
        k          TEXT NOT NULL,
        v          TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE (k, v)
    );
    CREATE INDEX IF NOT EXISTS idx_node_content ON nodes (content);
    CREATE INDEX IF NOT EXISTS idx_map_relation ON projection (relation_id);
    CREATE INDEX IF NOT EXISTS idx_map_node ON projection (node_id);
    CREATE INDEX IF NOT EXISTS idx_map_k ON identity (k);
`
	SqlReset = `
    TRUNCATE TABLE projection, relations, nodes CASCADE;
    ALTER SEQUENCE nodes_id_seq RESTART WITH 1;
    ALTER SEQUENCE relations_id_seq RESTART WITH 1;
    ALTER SEQUENCE projection_id_seq RESTART WITH 1;
`
)

// TODO 大重构
// Node -------------------------- 数据模型定义 --------------------------
// Node 对应nodes表
type Node struct {
	ID        int       `gorm:"column:id" db:"id"`
	Content   string    `gorm:"column:content" db:"content"`
	CreatedAt time.Time `gorm:"column:created_at" db:"created_at"`
}

type NodeType string

const (
	Trigger NodeType = "trigger"
	Result  NodeType = "result"
)

// Projection 对应projection表
type Projection struct {
	ID         int       `db:"id"`
	RelationID int       `db:"relation_id"`
	NodeID     int       `db:"node_id"`
	NodeType   NodeType  `db:"nodeType"`
	CreatedAt  time.Time `db:"created_at"`
}

// Relation 对应relations表
type Relation struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

// Identity 对应identity表
type Identity struct {
	ID        int       `db:"id"`
	K         string    `db:"k"`
	V         string    `db:"v"`
	CreatedAt time.Time `db:"created_at"`
}

// DBManager -------------------------- 数据库管理器 --------------------------
// DBManager 封装所有数据库操作
type DBManager struct {
	db *sql.DB
}

// NewDBManager 创建数据库管理器实例
// dsn格式: postgres://user:password@host:port/dbname?sslmode=disable
func NewDBManager(dsn string) (*DBManager, error) {
	// 1. 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 2. 配置连接池（生产环境关键）
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute) // 补充空闲超时，优化连接池

	// 3. 测试连接（确保连接有效）
	if err := db.Ping(); err != nil {
		db.Close() // 连接失败时关闭，避免资源泄露
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 4. 创建管理器实例
	dbManager := &DBManager{db: db}

	// 5. 初始化数据库表结构（建表/建索引/建枚举）
	if err := dbManager.initDB(); err != nil {
		db.Close() // 建表失败时关闭连接
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return dbManager, nil
}

// Close 关闭数据库连接
func (m *DBManager) Close() error {
	return m.db.Close()
}

// CreateNode -------------------------- Nodes 表操作 --------------------------
// CreateNode 创建节点（存在则返回现有记录，不存在则创建）
func (m *DBManager) CreateNode(content string) (*Node, error) {
	var node Node
	err := m.db.QueryRow(SqlNCreate, content).Scan(&node.ID, &node.Content, &node.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建节点失败: %w", err)
	}
	return &node, nil
}

// FindNode 根据ID查询节点
func (m *DBManager) FindNode(id int) (*Node, error) {
	var node Node
	err := m.db.QueryRow(SqlNFind, id).Scan(&node.ID, &node.Content, &node.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("节点不存在（ID:%d）: %w", id, err)
		}
		return nil, fmt.Errorf("查询节点失败: %w", err)
	}
	return &node, nil
}

// FindAllNodes 查询所有节点
func (m *DBManager) FindAllNodes() ([]Node, error) {
	rows, err := m.db.Query(SqlNAll)
	if err != nil {
		return nil, fmt.Errorf("查询所有节点失败: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.ID, &node.Content, &node.CreatedAt); err != nil {
			return nil, fmt.Errorf("解析节点数据失败: %w", err)
		}
		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历节点数据失败: %w", err)
	}

	return nodes, nil
}

// FindNodeByVal 根据内容查询节点
func (m *DBManager) FindNodeByVal(content string) (*Node, error) {
	var node Node
	err := m.db.QueryRow(SqlNFindByVal, content).Scan(&node.ID, &node.Content, &node.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("节点不存在（content:%s）: %w", content, err)
		}
		return nil, fmt.Errorf("查询节点失败: %w", err)
	}
	return &node, nil
}

// CreateProjection -------------------------- Projection 表操作 --------------------------
// CreateProjection 创建投影记录
func (m *DBManager) CreateProjection(relationID, nodeID int, nodeType NodeType) (*Projection, error) {
	var p Projection
	err := m.db.QueryRow(SqlPCreate, relationID, nodeID, nodeType).
		Scan(&p.ID, &p.RelationID, &p.NodeID, &p.NodeType, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建投影记录失败: %w", err)
	}
	return &p, nil
}

// DeleteProjection 根据ID删除投影记录
func (m *DBManager) DeleteProjection(id int) (*Projection, error) {
	var p Projection
	err := m.db.QueryRow(SqlPDelete, id).
		Scan(&p.ID, &p.RelationID, &p.NodeID, &p.NodeType, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("投影记录不存在（ID:%d）: %w", id, err)
		}
		return nil, fmt.Errorf("删除投影记录失败: %w", err)
	}
	return &p, nil
}

// FindProjectionByNode 根据节点ID查询投影记录
func (m *DBManager) FindProjectionByNode(nodeID int) ([]Projection, error) {
	rows, err := m.db.Query(SqlPFindByNode, nodeID)
	if err != nil {
		return nil, fmt.Errorf("查询投影记录失败: %w", err)
	}
	defer rows.Close()

	var projections []Projection
	for rows.Next() {
		var p Projection
		if err := rows.Scan(&p.ID, &p.RelationID, &p.NodeID, &p.NodeType, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("解析投影数据失败: %w", err)
		}
		projections = append(projections, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历投影数据失败: %w", err)
	}

	return projections, nil
}

// FindProjectionByRelation 根据关系ID查询投影记录
func (m *DBManager) FindProjectionByRelation(relationID int) ([]Projection, error) {
	rows, err := m.db.Query(SqlPFindByRelation, relationID)
	if err != nil {
		return nil, fmt.Errorf("查询投影记录失败: %w", err)
	}
	defer rows.Close()

	var projections []Projection
	for rows.Next() {
		var p Projection
		if err := rows.Scan(&p.ID, &p.RelationID, &p.NodeID, &p.NodeType, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("解析投影数据失败: %w", err)
		}
		projections = append(projections, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历投影数据失败: %w", err)
	}

	return projections, nil
}

// CreateRelation -------------------------- Relations 表操作 --------------------------
// CreateRelation 创建关系记录
func (m *DBManager) CreateRelation() (*Relation, error) {
	var r Relation
	err := m.db.QueryRow(SqlRCreate).Scan(&r.ID, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建关系记录失败: %w", err)
	}
	return &r, nil
}

// CreateIdentity -------------------------- Identity 表操作 --------------------------
// CreateIdentity 创建Identity记录
func (m *DBManager) CreateIdentity(k, v string) (*Identity, error) {
	var i Identity
	err := m.db.QueryRow(SqlICreate, k, v).Scan(&i.ID, &i.K, &i.V, &i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建Identity记录失败: %w", err)
	}
	return &i, nil
}

// FindIdentityByK 根据k查询Identity记录
func (m *DBManager) FindIdentityByK(k string) ([]Identity, error) {
	rows, err := m.db.Query(SqlIFindByK, k)
	if err != nil {
		return nil, fmt.Errorf("查询Identity记录失败: %w", err)
	}
	defer rows.Close()

	var identities []Identity
	for rows.Next() {
		var i Identity
		if err := rows.Scan(&i.ID, &i.K, &i.V, &i.CreatedAt); err != nil {
			return nil, fmt.Errorf("解析Identity数据失败: %w", err)
		}
		identities = append(identities, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历Identity数据失败: %w", err)
	}

	return identities, nil
}

// FindIdentityByKV 根据k和v查询Identity记录
func (m *DBManager) FindIdentityByKV(k, v string) (*Identity, error) {
	var i Identity
	err := m.db.QueryRow(SqlIFindByKv, k, v).Scan(&i.ID, &i.K, &i.V, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("identity记录不存在（k:%s, v:%s）: %w", k, v, err)
		}
		return nil, fmt.Errorf("查询Identity记录失败: %w", err)
	}
	return &i, nil
}

// UpdateIdentity 根据k更新对应的v值
func (m *DBManager) UpdateIdentity(k string, newV string) (*Identity, error) {
	// 1. 参数合法性校验
	if k == "" || newV == "" {
		return nil, errors.New("k和新的v值都不能为空")
	}

	// 2. 执行更新SQL
	var i Identity
	err := m.db.QueryRow(SqlIUpdate, k, newV).
		Scan(&i.ID, &i.K, &i.V, &i.CreatedAt)

	// 3. 错误处理
	if err != nil {
		// 记录不存在
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("要更新的identity记录不存在（k:%s）: %w", k, err)
		}
		// 唯一键冲突（k,v组合已存在）
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("更新失败：k=%s 对应的新v值=%s 已存在（k,v组合需唯一）: %w", k, newV, err)
		}
		return nil, fmt.Errorf("更新Identity记录失败: %w", err)
	}

	return &i, nil
}

// initDB -------------------------- 数据库初始化/重置 --------------------------
// initDB 初始化数据库（创建表、枚举、索引）
func (m *DBManager) initDB() error {
	_, err := m.db.Exec(SqlInit)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	return nil
}

// ResetDB 重置数据库（清空表数据，重置自增序列）
func (m *DBManager) ResetDB() error {
	_, err := m.db.Exec(SqlReset)
	if err != nil {
		return fmt.Errorf("重置数据库失败: %w", err)
	}
	return nil
}

var Db, err = NewDBManager(DefaultDSN)
