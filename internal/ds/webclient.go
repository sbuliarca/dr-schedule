package ds

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
)

var httpC = &http.Client{}

const DateFormat = "02.01.2006"

type PortalClient struct {
	/*"https://portal.lavitamed.ro/*/
	baseURL string
}

func NewPortalClient(baseURL string) PortalClient {
	return PortalClient{baseURL: baseURL}
}

func (c *PortalClient) GetFreeSlots(startDate time.Time, days int) ([]FreeDaySlots, error) {
	type FreeSlotsParams struct {
		Hcon            string `url:"Portal_HospitalConnectionValue,omitempty"`
		StartDate       string `url:"dataStart,omitempty"`
		EndDate         string `url:"dataEnd,omitempty"`
		Domain          int    `url:"idSpecialitate,omitempty"`
		IncludeMoreDays bool   `url:"maiMulteZile,omitempty"`
		DrCode          string `url:"medCod,omitempty"`
	}

	endDate := startDate.AddDate(0, 0, days)
	params := &FreeSlotsParams{Hcon: "ConnectionHM",
		StartDate:       startDate.Format(DateFormat),
		EndDate:         endDate.Format(DateFormat),
		Domain:          6,
		IncludeMoreDays: true,
		DrCode:          "AAT",
	}
	s := sling.New().Client(httpC).Base(c.baseURL).
		Get("DesktopModules/IWServices/API/Appointments/Programari_getIntervaleLibere").
		QueryStruct(params)
	var resp FreeSlotsResponse
	_, err := s.ReceiveSuccess(&resp)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving free slots, err: %w", err)
	}

	if resp.Type != "0" {
		return nil, fmt.Errorf("received non 0 type. Resp is %v", resp)
	}
	return resp.Items, nil
}

type FreeSlotsResponse struct {
	Type  string         `json:"TipRezultat"`
	Items []FreeDaySlots `json:"Items"`
}

type FreeDaySlots struct {
	Date   string `json:"DATA"`
	Hours  string `json:"ORE"`
	Doctor string `json:"MEDNUME"`
}

/*
{
    "Mesaj": "",
    "TipRezultat": "0",
    "NrTotalRezultate": "6",
    "Items": [
        {
            "ID": "1",
            "DATA": "04.01.2021",
            "ORE": "16:30,17:00,17:30,18:00,18:30,19:00,19:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "04.01.2021 16:30",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        },
        {
            "ID": "2",
            "DATA": "05.01.2021",
            "ORE": "08:00,08:30,09:00,09:30,10:00,10:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "05.01.2021 08:00",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        },
        {
            "ID": "3",
            "DATA": "06.01.2021",
            "ORE": "16:30,17:00,17:30,18:00,18:30,19:00,19:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "06.01.2021 16:30",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        },
        {
            "ID": "4",
            "DATA": "07.01.2021",
            "ORE": "08:00,08:30,09:00,09:30,10:00,10:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "07.01.2021 08:00",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        },
        {
            "ID": "5",
            "DATA": "08.01.2021",
            "ORE": "16:30,17:00,17:30,18:00,18:30,19:00,19:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "08.01.2021 16:30",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        },
        {
            "ID": "6",
            "DATA": "11.01.2021",
            "ORE": "16:30,17:00,17:30,18:00,18:30,19:00,19:30",
            "DEPFELCOD": "",
            "DEPFELNUME": "",
            "SPECIALITATEID": 6,
            "SPECIALITATENUME": "Gastroenterologie",
            "MEDCOD": "AAT",
            "MEDNUME": "Buliarca Alina",
            "DEPCOD": "AAL",
            "DEPNUME": "Gastroenterologie",
            "IDTIPORAR": null,
            "DENUMIRETIPORAR": "",
            "MIN_ORA": "11.01.2021 16:30",
            "AFISAREBILETTRIMITERE": true,
            "IDLOCATIE": null,
            "DENUMIRELOCATIE": "",
            "ADRESALOCATIE": "",
            "DESCRIERELOCATIE": ""
        }
    ]
}
*/
