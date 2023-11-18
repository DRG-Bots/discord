package main

type DeepDivesReqBody struct {
	StartTime string     `json:"startTime"`
	EndTime   string     `json:"endTime"`
	Variants  []DeepDive `json:"variants"`
}

type DeepDive struct {
	DDType string  `json:"type"`
	Name   string  `json:"name"`
	Biome  string  `json:"biome"`
	Seed   int     `json:"seed"`
	Stages []Stage `json:"stages"`
}

type Stage struct {
	Id        int    `json:"id"`
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Anomaly   string `json:"anomaly,omitempty"`
	Warning   string `json:"warning,omitempty"`
}
