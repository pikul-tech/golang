package main

import (
	//"encoding/json"
	"fmt"
)

// type Element struct {
// 	Actor    Actor  `json:"actor"`
// 	ID       string `json:"id"`
// 	Index    int    `json:"index"`
// 	Selected bool   `json:"selected"`
// 	Locked   bool   `json:"locked"`
// 	Hidden   bool   `json:"hidden"`
// 	Name     string `json:"name"`
// 	X        int    `json:"x"`
// 	Y        int    `json:"y"`
// 	EditedAt string `json:"editedAt"`
// 	Project  string `json:"project"`
// }

// type Actor struct {
// 	Type         string  `json:"type"`
// 	InstanceName string  `json:"instanceName"`
// 	X            int     `json:"x"`
// 	Y            int     `json:"y"`
// 	Origin       string  `json:"origin"`
// 	ScaleX       int     `json:"scaleX"`
// 	ScaleY       int     `json:"scaleY"`
// 	Rotation     int     `json:"rotation"`
// 	Width        int     `json:"width"`
// 	Height       int     `json:"height"`
// 	Xor          bool    `json:"xor"`
// 	Family       string  `json:"family"`
// 	LineHeight   int     `json:"lineHeight"`
// 	Size         int     `json:"size"`
// 	Align        string  `json:"align"`
// 	Text         string  `json:"text"`
// 	LockedWidth  bool    `json:"lockedWidth"`
// 	FixedWidth   int     `json:"fixedWidth"`
// 	ZPLFontID    string  `json:"zplFontId"`
// 	Baseline     string  `json:"baseline"`
// }

// func generateID() string {
// 	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
// 	b := make([]byte, 6)
// 	for i := range b {
// 		b[i] = chars[rand.Intn(len(chars))]
// 	}
// 	return string(b)
// }

// func main() {
// 	rand.Seed(time.Now().UnixNano())

// 	const project = "84d30031-0e9b-43c8-be5c-ee455c409172"

// 	elements := make([]Element, 0, 100)

// 	for i := 1; i <= 100; i++ {
// 		id := generateID()
// 		y := 100 + (i-1)*10

// 		element := Element{
// 			Actor: Actor{
// 				Type:         "zpltext",
// 				InstanceName: id,
// 				X:            100,
// 				Y:            y,
// 				Origin:       "tl",
// 				ScaleX:       1,
// 				ScaleY:       1,
// 				Rotation:     0,
// 				Width:        200,
// 				Height:       48,
// 				Xor:          false,
// 				Family:       "Arial",
// 				LineHeight:   0,
// 				Size:         16,
// 				Align:        "left",
// 				Text:         fmt.Sprintf("$.aso%d.nazwaglowna", i),
// 				LockedWidth:  true,
// 				FixedWidth:   200,
// 				ZPLFontID:    "0",
// 				Baseline:     "hanging",
// 			},
// 			ID:       id,
// 			Index:    i + 2,
// 			Selected: false,
// 			Locked:   false,
// 			Hidden:   false,
// 			Name:     "",
// 			X:        335,
// 			Y:        251,
// 			EditedAt: "2026-09-04T10:30:47.120Z",
// 			Project:  project,
// 		}

// 		elements = append(elements, element)
// 	}

// 	data, err := json.MarshalIndent(elements, "", "    ")
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(string(data))
// }

// func main() {
// 	const project = "84d30031-0e9b-43c8-be5c-ee455c409172"

// 	elements := generateElements(project)

// 	output, err := json.MarshalIndent(elements, "", "    ")
// 	if err != nil {
// 		panic(err)
// 	}

//		fmt.Println(string(output))
//	}
//
// []byte{0x1B, 0x21, 0x15, 0x0D, 0x0A},
func main() {

	zpl := "^XA^CI28" +
		"^FO28,28^GB770,417,3,B,0^FS" +
		"^FO24,361^A0N,30,30^FB296,1,0,L,0^FDTermometr pokojowy^FS" +
		"^FO28,61^A0N,30,30^FB296,1,0,L,0^FDJHM Mateusz Ratajczak^FS" +
		"^PQ1" +
		"^XZ"
	data := []byte(zpl)
	//for i := 0; i < 10; i++ {
	response, err := socketSendReceive("192.168.20.158:9100", data) //[]byte{0x7e, 0x21, 0x46})
	if err != nil {
		fmt.Println("Błąd:", err)
		return
	}
	fmt.Println("\r\nOdpowiedź: ", fmt.Sprintf("%s", response))
	//}
	//	fmt.Printf("Odpowiedź: % X\n", response)
}
