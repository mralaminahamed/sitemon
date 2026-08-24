package urlguard

import "testing"

func TestCheckBlocks(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1/x",
		"http://169.254.169.254/latest/meta-data/", // cloud metadata
		"http://10.0.0.5",
		"http://192.168.1.1",
		"http://[::1]",
		"http://0.0.0.0",
	}
	for _, u := range blocked {
		if err := Check(u); err == nil {
			t.Errorf("Check(%q) = nil, want blocked", u)
		}
	}
}

func TestCheckScheme(t *testing.T) {
	for _, u := range []string{"ftp://host/x", "file:///etc/passwd", "gopher://x"} {
		if err := Check(u); err == nil {
			t.Errorf("Check(%q) = nil, want scheme error", u)
		}
	}
}

func TestCheckAllowsPublic(t *testing.T) {
	// IP literal — LookupIP returns it without DNS, keeping the test offline.
	if err := Check("http://8.8.8.8/health"); err != nil {
		t.Errorf("Check(public) = %v, want nil", err)
	}
}

func TestAllowPrivateBypass(t *testing.T) {
	t.Setenv("ALLOW_PRIVATE_TARGETS", "true")
	if err := Check("http://127.0.0.1"); err != nil {
		t.Errorf("bypass: Check = %v, want nil", err)
	}
	// scheme is still validated under bypass
	if err := Check("ftp://127.0.0.1"); err == nil {
		t.Error("bypass should still reject bad scheme")
	}
}

func TestCheckBareHost(t *testing.T) {
	if err := Check("8.8.8.8"); err != nil {
		t.Errorf("bare public host: %v, want nil", err)
	}
	if err := Check("127.0.0.1"); err == nil {
		t.Error("bare loopback should be blocked")
	}
}
