package data

import (
	"goldenglow/node"
)

var (
	ExtraKeyWords = map[string]string{
		"A":         "[0xA]",
		"B":         "[0xB]",
		"C":         "[0xC]",
		"D":         "[0xD]",
		"E":         "[0xE]",
		"F":         "[0xF]",
		"he":        "[var_1]",
		"it":        "[var_2]",
		"[var_1]":   "someone",
		"[var_2]":   "something",
		"something": "[0xThing]",
		"someone":   "[0xPerson]",
	}
)

func InitMarks() {
	for k, v := range ExtraKeyWords {
		node.KVManager.Set(k, v)
	}
}

// if you lose sth and it is use less then sth you lose is useless
// if sth you lose is useless then you feel good
// 虽然出现2个连续的[0x],但是是作为R出现，T仍然合法,因此不会出现无法确定[0x]值的问题

//what is " says "
// says is the third-view single format of say
// if the mainc is third-view then say becomes says
// T:" says " means " say "
// R: [I] says & say (use protobuf)
