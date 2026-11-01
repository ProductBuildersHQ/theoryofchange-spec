package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/prismx"
)

func TestExportPRISM(t *testing.T) {
	doc, err := impactio.LoadDocument("../../examples/course-catalog/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range PRISMExportKinds {
		out, err := ExportPRISM(kind, doc, nil, nil)
		if err != nil || out.Kind != kind || out.Document == nil {
			t.Errorf("%s: %+v %v", kind, out, err)
		}
	}
	if _, err := ExportPRISM("mystery", doc, nil, nil); err == nil || !strings.Contains(err.Error(), "okr, journey, v2mom, dmaic") {
		t.Errorf("unknown kind: %v", err)
	}

	// A refused DMAIC export carries no document but still lists what it left out.
	noNumbers := *doc
	noNumbers.Indicators = append([]impact.Indicator(nil), doc.Indicators...)
	for i := range noNumbers.Indicators {
		noNumbers.Indicators[i].Baseline = nil
	}
	out, err := ExportPRISM("dmaic", &noNumbers, nil, nil)
	if !errors.Is(err, prismx.ErrNotDMAICReady) || out.Document != nil || len(out.Omissions) == 0 {
		t.Errorf("refused DMAIC export: %+v %v", out, err)
	}
}
