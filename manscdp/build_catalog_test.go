package manscdp

import (
	"strings"
	"testing"
)

func TestBuildCatalogResponse(t *testing.T) {
	items := []CatalogItem{
		{DeviceID: "34020000001310000001", Name: "大门", Manufacturer: "Hik", Model: "M1", Parental: 1, RegisterWay: 1, Status: "ON"},
	}
	out := BuildCatalogResponse("34020000002000000001", "9", items)
	for _, want := range []string{"<Response>", "<CmdType>Catalog</CmdType>", "<SN>9</SN>", "<SumNum>1</SumNum>", "<DeviceID>34020000001310000001</DeviceID>", "<Manufacturer>Hik</Manufacturer>", "<RegisterWay>1</RegisterWay>", "</DeviceList>"} {
		if !strings.Contains(out, want) {
			t.Fatalf("catalog response missing %q:\n%s", want, out)
		}
	}
}

func TestBuildCatalogNotifyIncludesItemFields(t *testing.T) {
	out := BuildCatalogNotify("34020000002000000001", "1", []CatalogItem{{DeviceID: "ch1", Name: "n", Model: "M"}})
	if !strings.Contains(out, "<Notify>") || !strings.Contains(out, "<Model>M</Model>") {
		t.Fatalf("unexpected notify:\n%s", out)
	}
}
