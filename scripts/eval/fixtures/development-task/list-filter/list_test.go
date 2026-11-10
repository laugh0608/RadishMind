package tasklist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestTaskList(t *testing.T) {
	for _, tc := range []struct {
		name   string
		query  string
		tasks  []Task
		status int
		ids    []string
	}{
		{"all", "", []Task{{"a", "open"}, {"b", "done"}, {"c", "open"}}, 200, []string{"a", "b", "c"}},
		{"filter_before_limit", "?state=open&limit=1", []Task{{"a", "open"}, {"b", "done"}, {"c", "open"}}, 200, []string{"a"}},
		{"limit_two", "?limit=2", []Task{{"a", "open"}, {"b", "done"}, {"c", "open"}}, 200, []string{"a", "b"}},
		{"exact_limit", "?limit=1", []Task{{"a", "open"}}, 200, []string{"a"}},
		{"empty", "?state=done", []Task{{"a", "open"}}, 200, []string{}},
		{"invalid_state", "?state=unknown", nil, 400, nil},
		{"zero_limit", "?limit=0", nil, 400, nil},
		{"invalid_limit", "?limit=one", nil, 400, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Handler(tc.tasks).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks"+tc.query, nil))
			if w.Code != tc.status {
				t.Fatalf("status=%d want=%d", w.Code, tc.status)
			}
			if tc.status != http.StatusOK {
				return
			}
			var got []Task
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got == nil {
				t.Fatalf("expected JSON array: %s, error=%v", w.Body, err)
			}
			ids := make([]string, 0, len(got))
			for _, task := range got {
				ids = append(ids, task.ID)
			}
			if !reflect.DeepEqual(ids, tc.ids) {
				t.Fatalf("ids=%v want=%v", ids, tc.ids)
			}
		})
	}
}
