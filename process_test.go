package RoutingA

import "testing"

func parseOne(t *testing.T, program string) Routing {
	t.Helper()
	rules, err := Parse(program)
	if err != nil {
		t.Fatalf("%s: %v", program, err)
	}
	if len(rules) != 1 {
		t.Fatalf("%s: %d rules", program, len(rules))
	}
	r, ok := rules[0].(Routing)
	if !ok {
		t.Fatalf("%s: not a routing rule: %#v", program, rules[0])
	}
	return r
}

func processOf(t *testing.T, program string) ProcessMatcher {
	t.Helper()
	r := parseOne(t, program)
	if len(r.And) != 1 {
		t.Fatalf("%s: %d conditions", program, len(r.And))
	}
	m, isProcess, err := r.And[0].ProcessMatcher()
	if err != nil {
		t.Fatalf("%s: %v", program, err)
	}
	if !isProcess {
		t.Fatalf("%s: %q is not process(...)", program, r.And[0].Name)
	}
	return m
}

// A bare value is a name: it is what a rule author writes first.
func TestProcessBareParameterIsAName(t *testing.T) {
	m := processOf(t, `process(firefox)->proxy`)
	if len(m.Names) != 1 || m.Names[0] != "firefox" {
		t.Fatalf("%+v", m)
	}
	if len(m.Paths) != 0 || len(m.Folders) != 0 {
		t.Fatalf("%+v", m)
	}
}

func TestProcessNamedParameters(t *testing.T) {
	m := processOf(t, `process(name: "fire*", path: /usr/lib/firefox/firefox, folder: /opt/apps)->proxy`)
	if len(m.Names) != 1 || m.Names[0] != "fire*" {
		t.Fatalf("names: %+v", m.Names)
	}
	if len(m.Paths) != 1 || m.Paths[0] != "/usr/lib/firefox/firefox" {
		t.Fatalf("paths: %+v", m.Paths)
	}
	if len(m.Folders) != 1 || m.Folders[0] != "/opt/apps" {
		t.Fatalf("folders: %+v", m.Folders)
	}
}

// A process(...) rule is an ordinary routing rule: it keeps its other
// conditions and its outbound.
func TestProcessCombinesWithOtherConditions(t *testing.T) {
	r := parseOne(t, `process(firefox) && port(443) && network(tcp)->proxy`)
	if len(r.And) != 3 {
		t.Fatalf("%d conditions", len(r.And))
	}
	if r.Out != "proxy" {
		t.Fatalf("outbound %q", r.Out)
	}
	if _, isProcess, err := r.And[0].ProcessMatcher(); !isProcess || err != nil {
		t.Fatalf("the first condition is not process(...): %+v %v", r.And[0], err)
	}
}

func TestProcessMatcherIsNotConfusedWithOtherFunctions(t *testing.T) {
	for _, name := range []string{"domain", "ip", "port", "network", "protocol", "inboundTag"} {
		r := parseOne(t, name+"(x)->proxy")
		if _, isProcess, err := r.And[0].ProcessMatcher(); isProcess || err != nil {
			t.Fatalf("%s: isProcess=%v err=%v", name, isProcess, err)
		}
	}
}

// An unknown parameter is refused: the rule cannot be enforced, and dropping
// the key would widen what the rule matches.
func TestProcessRefusesAnUnknownParameter(t *testing.T) {
	r := parseOne(t, `process(uid: 1000)->proxy`)
	if _, isProcess, err := r.And[0].ProcessMatcher(); !isProcess || err == nil {
		t.Fatalf("isProcess=%v err=%v", isProcess, err)
	}
}

func TestProcessRefusesAnEmptyCondition(t *testing.T) {
	r := parseOne(t, `process()->proxy`)
	if _, isProcess, err := r.And[0].ProcessMatcher(); !isProcess || err == nil {
		t.Fatalf("isProcess=%v err=%v", isProcess, err)
	}
}

func TestProcessMatch(t *testing.T) {
	m := ProcessMatcher{
		Names:   []string{"firefox", "chrom*"},
		Paths:   []string{"/usr/bin/curl"},
		Folders: []string{"/opt/apps/"},
	}
	for _, tc := range []struct {
		name, path string
		want       bool
	}{
		{"firefox", "/usr/lib/firefox/firefox", true},
		{"chrome", "/opt/google/chrome/chrome", true},
		{"curl", "/usr/bin/curl", true},
		{"curl", "/usr/local/bin/curl", false},
		{"tool", "/opt/apps/tool/bin/tool", true},
		{"tool", "/opt/apps-other/tool", false},
		{"bash", "/bin/bash", false},
	} {
		if got := m.Match(tc.name, tc.path); got != tc.want {
			t.Errorf("%s (%s) = %v, want %v", tc.name, tc.path, got, tc.want)
		}
	}
}

// A malformed pattern must not turn into "match everything".
func TestProcessMalformedPatternMatchesNothing(t *testing.T) {
	m := ProcessMatcher{Names: []string{"["}}
	if m.Match("firefox", "/usr/bin/firefox") {
		t.Fatal("a malformed glob matched")
	}
}

func TestWithinFolder(t *testing.T) {
	for _, tc := range []struct {
		dir, folder string
		want        bool
	}{
		{"/opt/app", "/opt/app", true},
		{"/opt/app/bin", "/opt/app", true},
		{"/opt/application", "/opt/app", false},
		{"/opt/app", "/", true},
	} {
		if got := withinFolder(tc.dir, tc.folder); got != tc.want {
			t.Errorf("withinFolder(%q, %q) = %v, want %v", tc.dir, tc.folder, got, tc.want)
		}
	}
}