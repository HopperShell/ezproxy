package configurator

import "testing"

func TestSystemCAName(t *testing.T) {
	s := &SystemCA{}
	if s.Name() != "system_ca" {
		t.Errorf("Name() = %q", s.Name())
	}
}
