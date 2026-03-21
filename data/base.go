package data

type NV []string
type TR struct {
	tv NV
	rv NV
}

var (
	base = map[string]TR{
		"say-trigger": {
			tv: NV{"[say] [0x01] to [0x02] : [0x03]"},
			rv: NV{"[0x01] say [0x03] to [0x02]"},
		},
		"make KV CLi": {
			tv: NV{"[say] [0xb] to [GG] : [0x01] is [0x02]"},
			rv: NV{"[0x01] is [0x02]", "[BetterSpeaker] [I] [0x01] -> [0x02] & [Speak] & [0xb]"},
		},
		"kv": {
			tv: NV{"[0x01] is [0x02]"},
			rv: NV{"[I] [0x01] & [0x02]"},
		},
		"check KV CLi": {
			tv: NV{"[say] [0xb] to [GG] : [IC] [0x01] -> [0x02]", "[check] [0x02] & [0x01] & [IC]"},
			rv: NV{"[BetterSpeaker] yes & [Speak] & [0xb]"},
		},
		"check KV Not CLi": {
			tv: NV{"[say] [0xb] to [GG] : [IC] [0x01] -> [0x02]", "[check] [0x02] & [0x01] & [ICN]"},
			rv: NV{"[BetterSpeaker] no & [Speak] & [0xb]"},
		},
		"get KV CLi": {
			tv: NV{"[say] [0xb] to [GG] : ? [0x01]"},
			rv: NV{"[0xb] to [GG] : ? [0x01]"},
		},
		"get KV 0": {
			tv: NV{"[0xb] to [GG] : ? [0x01]"},
			rv: NV{"[BetterSpeaker] [0x02] & [Speak] & [0xb]", "[0x02] : {[0x01]} //"},
		},
		// "get S&L 0": {
		// 	tv: NV{"[say] [0x01] to [0xg] : [0x03]"},
		// 	rv: NV{"[I] [speaker] & [0x01] & [PB]", "[I] [listener] & [0xg] & [PB]"},
		// },
		"if then": {
			tv: NV{"[say] [0x01] to [GG] : if [0x03] then [0x04]"},
			rv: NV{
				"[P] [0x04] [R] | [0x03]",
			},
		},
		// "get KV 2": {
		// 	tv: NV{"[say] [0xb] to [GG] : what is [0x01]"},
		// 	rv: NV{"[0xb] to [GG] : ? [0x01]"},
		// },
		// "make KV 3": {
		// 	tv: NV{"[say] [0xb] to [GG] : [0x01] means [0x02]"},
		// 	rv: NV{"[say] [0xb] to [GG] : [I] [0x01] -> [0x02]"},
		// },
		// "check KV name": {
		// 	tv: NV{"[say] [0xb] to [GG] : is [0x01] name [0x02]"},
		// 	rv: NV{"[say] [0xb] to [GG] : [IC] [0x03] -> [0x02]", "[0x03] : [0x01] name //"},
		// },
		"I ask": {
			tv: NV{"[I Ask] [0x01]"},
			rv: NV{"[GG] don't know what [0x01] is"},
		},
	}
)
