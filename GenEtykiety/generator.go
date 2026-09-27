package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

var project = "84d30031-0e9b-43c8-be5c-ee455c409172"

var dataIDs = []string{
	"403106_OKL",
	"403107_OKL",
	"403108_OKL",
	"403109_OKL",
	"403110_OKL",
}

const zplDataTemplate = `{
	"actor": {
		"type": "zpldata",
		"instanceName": "aso1",
		"x": 0,
		"y": 0,
		"origin": "tl",
		"scaleX": 1,
		"scaleY": 1,
		"rotation": 0,
		"width": 40,
		"height": 40,
		"xor": false,
		"sourceKey": "asortyment",
		"dataId": "403106_OKL",
		"data": {
			"_abc_historia": null,
			"_abc_klasyfikacja": null,
			"_abc_klasyfikacja_w_linii": null,
			"_abc_odkiedy": null,
			"_abc_poprzednia": null,
			"_abc_propozycja_pm": null,
			"_categorymanager": "Cybulska Olga",
			"_ceny": null,
			"_dostawca_kraj": null,
			"_dostepne": 0,
			"_dostepnejednostki": [
				{
					"id": "OPJ",
					"name": "op. jedn."
				}
			],
			"_ilosclogistyk": 0,
			"_iloscnapaleciezakupy": null,
			"_jednostkamiary": "szt.",
			"_jednostkamiary_de": "Stk.",
			"_jednostkamiary_en": "pcs",
			"_jednostkamiary_vendo": "szt.",
			"_jednostkamiarymagazyn": null,
			"_jednostkamiarysprzedaz": null,
			"_kategoria": null,
			"_krajpochodzenia": null,
			"_krajpochodzenia_de": null,
			"_krajpochodzenia_en": null,
			"_nadkategoria": null,
			"_nazwa_de": null,
			"_nazwa_en": null,
			"_nazwa_ru": null,
			"_opiekun": "Wołosz Mateusz",
			"_opiekun_planowanie": null,
			"_opiekun_rd_bio": null,
			"_opiekun_zakupy": null,
			"_ostatnia_zmiana": null,
			"_podkategoria": null,
			"_siec": null,
			"_status": "Inny",
			"_status_kolor": "#DCBBF3",
			"_stawka_vat": "23%",
			"_sur_historia": null,
			"_sur_klasyfikacja": null,
			"_sur_odkiedy": null,
			"_sur_poprzednia": null,
			"_symbol_grupy_raportowej": null,
			"_vat_wartosc": 23,
			"_warunki_przechowywania": null,
			"_warunki_przechowywania_de": null,
			"_warunki_przechowywania_en": null,
			"_zakup_historia": null,
			"_zakup_klasyfikacja": null,
			"_zakup_odkiedy": null,
			"_zakup_poprzednia": null,
			"_zapas_komentarz": null,
			"_zapas_max": null,
			"_zapas_min": null,
			"akceptacjaksiegowosc": true,
			"categorymanager": "16922635.BRO.1",
			"czescizamienne": null,
			"czesczamienna": false,
			"czy_produkt_strategiczny": false,
			"data_akceptacjaksiegowosc": "2025-06-10 14:59",
			"data_wycofania": null,
			"datanowosci": null,
			"datawyprzedazy": null,
			"dc": "2025-06-10T15:00:50",
			"del": false,
			"ean13_blad": null,
			"ean13_opj": "10035375",
			"forma_wtryskowa": null,
			"glebokosc_opj": null,
			"glebokosc_opm": null,
			"glebokosc_opz": null,
			"grupa": null,
			"grupa_ciepla": null,
			"grupa_zimna": null,
			"haccp": false,
			"id": "403106_OKL",
			"innyopiekun": true,
			"jednostka_opj": "szt.",
			"jednostkamiary": "SZT",
			"jestczesciazamienna": null,
			"kanibal": null,
			"kanibalizowane": null,
			"kategoria": null,
			"kodkreskowy": null,
			"kupowany": false,
			"logistykadomyslna": null,
			"marka": null,
			"masa_opj": null,
			"masa_opm": null,
			"masa_opz": null,
			"material": null,
			"matmarketingowy": false,
			"mid": "4509393860.BRO.1",
			"nadawaniegtin": "BROWIN",
			"nadkategoria": null,
			"nazwa": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nazwa_wewnetrzna": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nazwaglowna": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nazwanafakture": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nazwanapalmtopa": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nazwawww": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"nowagrupaasortymentowa": null,
			"nrpartii": true,
			"objetosctransportowa": null,
			"ok": true,
			"opakowaniemaster": null,
			"opakowaniezbiorcze": null,
			"opiekun": "3276325960.BRO.1",
			"opiekun_planowanie": null,
			"opiekun_rd_bio": null,
			"opiekun_zakupy": null,
			"ostatnie5znakowean": "35375",
			"pc": "ALCZE",
			"podkategoria": null,
			"podlega_lsse": false,
			"podzialksiegowy": "MATERIAL",
			"polprodukt": true,
			"posiadadatewaznosci": true,
			"pozostaladzialalnosc": false,
			"produkowany": true,
			"prognoza_nowosci": null,
			"propozycja_pm": null,
			"rodzaj": "MAGAZ",
			"rodzajpalety1": "EPAL",
			"siec": null,
			"slug": "drozdze-turbo-x-pure-21-3-360g-100l-oklejanie",
			"status": "INNY",
			"statuszapasu": null,
			"surowiec": null,
			"symbol": "403106_OKL",
			"symbol_bez_zer": "403106_OKL",
			"symboldokomunikacji": "403106_OKL",
			"symbolroboczy": "403106_OKL",
			"sync": {
				"stary": true,
				"vendo": true,
				"wms": true
			},
			"sync_ids": "(,15193,,,,,)",
			"szczegolowakj": true,
			"szerokosc_opj": null,
			"szerokosc_opm": null,
			"szerokosc_opz": null,
			"title": "Drożdże Turbo X-Pure 21,3% 360g 100L - oklejanie",
			"towarhandlowy": false,
			"tylko_pelne_opakowania": false,
			"usluga": false,
			"vat": "23",
			"vendo": {
				"uzyj_daty": true,
				"uzyj_numeru": true
			},
			"vendostatus": "UPD",
			"widocznosc": null,
			"wmsstatus": "UPD",
			"wysokosc_opj": null,
			"wysokosc_opm": null,
			"wysokosc_opz": null,
			"zawartosc_opj": 1,
			"zawartosc_opm": null,
			"zawartosc_opz": null,
			"zdjecie": null,
			"zestaw": false
		}
	},
	"id": "9jt8ab",
	"index": 2,
	"selected": false,
	"locked": false,
	"hidden": false,
	"name": "",
	"x": 202,
	"y": -85,
	"editedAt": "2026-09-04T10:30:47.120Z",
	"project": "84d30031-0e9b-43c8-be5c-ee455c409172"
}`

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func generateElements(project string) []map[string]interface{} {
	elements := make([]map[string]interface{}, 0, len(dataIDs)*2)

	// ============================================================
	// ZPLDATA
	// ============================================================
	for i, dataID := range dataIDs {
		aso := i + 1

		var element map[string]interface{}
		if err := json.Unmarshal([]byte(zplDataTemplate), &element); err != nil {
			panic(err)
		}

		actor := element["actor"].(map[string]interface{})
		data := actor["data"].(map[string]interface{})

		actor["instanceName"] = fmt.Sprintf("aso%d", aso)
		actor["origin"] = fmt.Sprintf("tl%d", aso)
		actor["dataId"] = dataID

		data["id"] = dataID
		data["symbol"] = dataID
		data["symbol_bez_zer"] = dataID
		data["symboldokomunikacji"] = dataID
		data["symbolroboczy"] = dataID

		element["id"] = generateID()
		element["index"] = i + 2
		element["project"] = project

		elements = append(elements, element)
	}

	// ============================================================
	// ZPLTEXT
	// ============================================================
	for i := range dataIDs {
		aso := i + 1
		id := generateID()

		element := map[string]interface{}{
			"actor": map[string]interface{}{
				"type":         "zpltext",
				"instanceName": id,
				"x":            100,
				"y":            100 + i*10,
				"origin":       fmt.Sprintf("tl%d", aso),
				"scaleX":       1,
				"scaleY":       1,
				"rotation":     0,
				"width":        200,
				"height":       48,
				"xor":          false,
				"family":       "Arial",
				"lineHeight":   0,
				"size":         16,
				"align":        "left",
				"text":         fmt.Sprintf("$.aso%d.nazwaglowna", aso),
				"lockedWidth":  true,
				"fixedWidth":   200,
				"zplFontId":    "0",
				"baseline":     "hanging",
			},
			"id":       id,
			"index":    len(dataIDs) + i + 2,
			"selected": false,
			"locked":   false,
			"hidden":   false,
			"name":     "",
			"x":        335,
			"y":        251,
			"editedAt": "2026-09-04T10:30:47.120Z",
			"project":  project,
		}

		elements = append(elements, element)
	}

	return elements
}
