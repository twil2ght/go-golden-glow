1. pronoun的cache需要完善
2. 已经学习过的res-cond组合，下次碰到肯定不用在学习一遍而是增加一个learned hook
3. 学习具体动词的具体用法后要删除不对的模板
4. double_noun:让前面的noun具有adj词性
5. 多个of连接的noun要从第一个of看起，在分析时加上条件：是否为第一个prep
6. 对于入口的[legal] C say : A.采用TODO的方法依次检测各种句型而不是一口气全部检测一遍
7. 需要清理Match:can等临时内容(先删除再进行GA)
8. cited的3种情况也要递进式的进行
9. 创建一个总的GA（像GA：VP)检测conj,S+VP,VP,O
10. 给ticker设置灵活计时时间(参数化)
11. 将CR的保存形式改成普遍的存储形式,其中C必须是动名词句子；R必须是that从句
12. apple is a noun -> apple 's being a noun
13. method:1)主语变成形容词性;2)动词变ing
14. 自动生成句子模板
15. wh- 句子
16. adv句子
17. and or句子
18. that clause的Resolver配置
19. when you have more than one words that you don't ensure,say them out one by one
20. executor的steps设计不合理完全可以移除然后使用正常的if-then
21. if I say : do something;then you try to do something(用于before doing sth加行动前检测)
22. learning中的VP会触发$increase-apple
23. 只让第一个Cond使用template，其他的Cond用template也没用还是需要第一个Cond触发才有用
