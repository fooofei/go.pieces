package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

// A biz test

type SdkTerm struct {
	Uuid  string   `json:"sdkUuid"`
	SrcIp []string `json:"srcIp,omitempty"`
}

type App struct {
	KeyId    string    `json:"keyId"`
	SdkCnt   uint64    `json:"onlineSdkCount"`
	IBytes   uint64    `json:"inPayloadLength"`
	OBytes   uint64    `json:"outPayloadLength"`
	DecTerms []SdkTerm `json:"decreaseSdkInfo"`
	IncTerm  []SdkTerm `json:"increaseSdkInfo"`
}

type AppSet map[string]*App

func NewAppSet() AppSet {
	return make(map[string]*App, 0)
}

func (as AppSet) Add(apps map[string]interface{}) error {
	stats, is := apps["statistic"].([]interface{})
	if !is {
		return fmt.Errorf("not contain statistic array")
	}
	for idx := range stats {
		sapp := new(App)
		appb, _ := json.Marshal(stats[idx])
		err := json.Unmarshal(appb, sapp)
		if err != nil {
			return err
		}
		if as[sapp.KeyId] == nil {
			as[sapp.KeyId] = new(App)
		}
		dapp := as[sapp.KeyId]
		dapp.IncTerm = append(dapp.IncTerm, sapp.IncTerm...)
		dapp.IBytes = sapp.IBytes
		dapp.OBytes = sapp.OBytes
		dapp.SdkCnt = sapp.SdkCnt

	}
	return nil
}

func (as AppSet) Sub(apps map[string]interface{}) error {
	stats, _ := apps["statistic"].([]interface{})
	for idx := range stats {
		sapp := new(App)
		appb, _ := json.Marshal(stats[idx])
		err := json.Unmarshal(appb, &sapp)
		if err != nil {
			return err
		}
		if as[sapp.KeyId] == nil {
			as[sapp.KeyId] = new(App)
		}
		dapp := as[sapp.KeyId]
		dapp.DecTerms = append(dapp.DecTerms, sapp.DecTerms...)
		dapp.IBytes = sapp.IBytes
		dapp.OBytes = sapp.OBytes
		dapp.SdkCnt = sapp.SdkCnt
	}
	return nil
}

func TestCltMsg1(t *testing.T) {

	var err error

	addTerm := `{
        "statistic": [
            {
                "keyId": "6ba724d5313044949eef7d25fae42b52",
                "onlineSdkCount": 1,
                "inPayloadLength": 2,
                "outPayloadLength": 3,
                "decreaseSdkInfo": [],
                "increaseSdkInfo": [
                    {
                        "sdkUuid": "1234567890abcdef1234567890abcdefffff",
                        "srcIp": [
                            "10.177.96.185"
                        ]
                    }
                ]
            }
        ],
        "wkrIdx": 0,
        "rsn": "addTerm"
    }`

	subTerm := `{
        "statistic": [
            {
                "keyId": "6ba724d5313044949eef7d25fae42b52",
                "onlineSdkCount": 0,
                "inPayloadLength": 0,
                "outPayloadLength": 0,
                "decreaseSdkInfo": [  {
                        "sdkUuid": "1234567890abcdef1234567890abcdefffff"
                        
                    }],
                "increaseSdkInfo": [
                  
                ]
            }
        ],
        "wkrIdx": 0,
        "rsn": "subTerm"
    }`

	addJ := make(map[string]interface{})
	err = json.Unmarshal([]byte(addTerm), &addJ)
	if err != nil {
		t.Fatal(err)
	}
	subJ := make(map[string]interface{})
	err = json.Unmarshal([]byte(subTerm), &subJ)
	if err != nil {
		t.Fatal(err)
	}

	dst := NewAppSet()

	dst.Add(addJ)
	dst.Sub(subJ)

	dstj, _ := json.MarshalIndent(dst, "", "  ")

	assert.Equal(t, string(dstj), `{
  "6ba724d5313044949eef7d25fae42b52": {
    "keyId": "",
    "onlineSdkCount": 0,
    "inPayloadLength": 0,
    "outPayloadLength": 0,
    "decreaseSdkInfo": [
      {
        "sdkUuid": "1234567890abcdef1234567890abcdefffff"
      }
    ],
    "increaseSdkInfo": [
      {
        "sdkUuid": "1234567890abcdef1234567890abcdefffff",
        "srcIp": [
          "10.177.96.185"
        ]
      }
    ]
  }
}`)
}

func TestSdkTerm1(t *testing.T) {
	term := &SdkTerm{}
	//term.SrcIp = make([]string,0)
	term.Uuid = "123455"

	j, _ := json.Marshal(term)

	assert.Equal(t, string(j), `{"sdkUuid":"123455"}`)
}
