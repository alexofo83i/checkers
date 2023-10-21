package main

func convertWhoDo2WhoDoNext(whodo byte) byte {
	var whodonext byte
	if whodo == black {
		whodonext = white
	} else {
		whodonext = black
	}
	return whodonext
}

func (state *Table) convertGame2PlayState() *PlayState {
	var whodo byte
	if state.Whodo == "checker-black" {
		whodo = black
	} else {
		whodo = white
	}
	playStateInit := NewPlayState(whodo, 24)
	// convert Game.State aka Table to PlayState
	for _, tr := range state.Trs {
		for _, td := range tr.Tds {
			if td.Ch.Id != "" {
				playStateInit.f2c[td.Id] = td.Ch.Id
				playStateInit.c2f[td.Ch.Id] = td.Id
			}
		}
	}
	return playStateInit
}

func (playState *PlayState) convertPlayState2Table(name string) *Table {
	table := initTable(nil)
	if playState.whodo == black {
		table.Whodo = "checkers-black"
	} else {
		table.Whodo = "checkers-white"
	}
	table.Name = name
	for i, tr := range table.Trs {
		for j, td := range tr.Tds {
			chid := playState.f2c[td.Id]
			if chid != "" {
				table.Trs[i].Tds[j].Ch = Checker{
					Id: chid,
				}
				//log.Default().Println(td.Id, " =", chid)
			}
		}
	}
	return table
}
