package source

import "strings"

type Refer struct {
	UUKey string `json:"uukey"`
	KeyBy string `json:"keyby"`
	TxtBy string `json:"txtby"`
	Image string `json:"image"`
	Using string `json:"using"`
}

func (self *Refer) Parse(using string) *Refer {
	parts := strings.Split(using, "@")
	if len(parts) == 1 {
		self.Using = using
		return self
	}
	fields := parts[0]
	self.Using = parts[1]
	parts = strings.Split(fields, ",")
	if len(parts) == 1 {
		self.KeyBy = "uukey"
		self.TxtBy = parts[0]
	} else {
		self.TxtBy = parts[1]
		self.KeyBy = parts[0]
	}
	return self
}

func (self *Refer) Load(object map[string]string) *Refer {
	if using, ok := object["using"]; ok {
		self.Using = using
	}
	if image, ok := object["image"]; ok {
		self.Image = image
	}
	if keyby, ok := object["keyby"]; ok {
		self.KeyBy = keyby
	} else {
		self.KeyBy = "uukey"
	}
	if txtby, ok := object["txtby"]; ok {
		self.TxtBy = txtby
	} else {
		self.TxtBy = "label"
	}
	return self
}
