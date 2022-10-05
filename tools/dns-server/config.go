package main

type Record struct {
	Domain string
	Host   string
}

type Config struct {
	Addr    string
	Net     string
	Records []*Record
}

func (c *Config) ToMap() map[string]*Record {
	var mapRecords = make(map[string]*Record, 0)
	for _, rec := range c.Records {
		mapRecords[rec.Domain] = rec
	}
	return mapRecords
}
