package service

import "testing"

func TestCheckDNS(t *testing.T) {
	// .invalid is reserved (RFC 6761); resolvers should not return A/AAAA
	// records, unlike arbitrary labels that some DNS may resolve or sinkhole.
	s1 := &Service{
		Name: "nonexistent-reserved.invalid",
	}
	_, err := s1.CheckDNS()
	if err == nil {
		t.Fatal("should be error")
	}

	s2 := &Service{
		Name: "baidu.com",
	}
	_, err = s2.CheckDNS()
	if err != nil {
		t.Fatal(err)
	}
}
