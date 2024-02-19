package wpamelia

import (
	"encoding/json"
	"fmt"
	"github.com/dr-schedule/internal/ds"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

var httpC = &http.Client{}

type Client struct {
	/*"https://lavitamed.ro/*/
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

var slotParseFormat = "2006-01-02T15:04"

func (c *Client) GetBusySlots() (ds.Slots, error) {
	url := c.baseURL + "wp-admin/admin-ajax.php?action=wpamelia_api&call=/slots&monthsLoad=1&serviceId=13&serviceDuration=1800&providerIds=23&group=1&page=booking&persons=1"
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed making get request: %w", err)
	}
	defer r.Body.Close()

	code := r.StatusCode
	if code != 200 {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed reading whole body: %w", err)
		}

		return nil, fmt.Errorf("get free slots: didn't receive expected status, but: %d, body: %s", code, body)
	}
	var resp SlotsResponse

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("failed decoding to json, err: %w", err)
	}

	logrus.Debugf("got WP slots response %+v", resp)

	if !strings.Contains(resp.Message, "Successfully") {
		return nil, fmt.Errorf("filed retrieving free slots, unexpected message: %+v", resp)
	}

	return toCommonResponse(resp)
}

func toCommonResponse(resp SlotsResponse) (ds.Slots, error) {
	res := make(ds.Slots)

	for day, busySlots := range resp.Data.Occupied {
		for slotTime := range busySlots {
			fullSlotTime := fmt.Sprintf("%sT%s", day, slotTime)
			slotStart, err := time.ParseInLocation(slotParseFormat, fullSlotTime, ds.LocalLoc)
			if err != nil {
				return nil, fmt.Errorf("failed parsing busy full slot: %w", err)
			}

			res[slotStart.Unix()] = struct{}{}
		}
	}
	return res, nil
}

type SlotsResponse struct {
	Message string `json:"message"`
	Data    struct {
		Minimum  string              `json:"minimum"`
		Maximum  string              `json:"maximum"`
		Occupied map[string]BusySlot `json:"occupied"`
	} `json:"data"`
}

type BusySlot map[string]json.RawMessage

/*
{
    "message": "Successfully retrieved free slots",
    "data": {
        "minimum": "2024-02-18 06:51",
        "maximum": "2024-09-05 06:51",
        "slots": {
            "2024-02-19": {
                "18:30": [
                    [
                        23,
                        null
                    ]
                ],
                "19:00": [
                    [
                        23,
                        null
                    ]
                ]
            },
            "2024-02-20": {
                "09:30": [
                    [
                        23,
                        null
                    ]
                ],
                "10:00": [
                    [
                        23,
                        null
                    ]
                ]
            },
        },
        "occupied": {
            "2024-02-19": {
                "17:30": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ],
                "18:00": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ]
            },
            "2024-02-20": {
                "07:30": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ],
                "08:00": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ],
                "08:30": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ],
                "09:00": [
                    [
                        23,
                        null,
                        0,
                        13
                    ]
                ]
            },
        },
        "busyness": {
            "2024-02-19": 50,
            "2024-02-20": 67,
            "2024-02-21": 25,
            "2024-02-23": 0,
            "2024-02-26": 25,
            "2024-02-27": 67,
            "2024-02-28": 75,
            "2024-02-29": 0,
            "2024-03-01": 0,
            "2024-03-04": 0,
            "2024-03-05": 0,
            "2024-03-06": 0,
            "2024-03-07": 0,
            "2024-03-08": 0,
            "2024-03-11": 0,
            "2024-03-12": 0
        },
        "lastProvider": null,
        "appCount": {
            "23": {
                "2024-02-19": 0,
                "2024-02-20": 0,
                "2024-02-21": 0,
                "2024-02-22": 0,
                "2024-02-23": 0,
                "2024-02-26": 0,
                "2024-02-27": 0,
                "2024-02-28": 0,
                "2024-02-29": 0,
                "2024-03-01": 0,
                "2024-03-04": 0,
                "2024-03-05": 0,
                "2024-03-06": 0,
                "2024-03-07": 0,
                "2024-03-08": 0,
                "2024-03-11": 0,
                "2024-03-12": 0
            }
        }
    }
}
*/
