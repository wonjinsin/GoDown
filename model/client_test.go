package model

import "testing"

func TestGetDomainFromURL(t *testing.T) {

	if d := getDomainFromURL("https://www.test.kr/test/test123"); d != "https://www.test.kr" {
		t.Errorf("https Domain is not valid %s", d)
	}

	if d := getDomainFromURL("http://www.test.kr/test/test123"); d != "http://www.test.kr" {
		t.Errorf("http Domain is not valid %s", d)
	}
}
