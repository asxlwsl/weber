package router

import (
	"testing"
)

func TestRouter(t *testing.T) {

	type TestCase struct {
		name       string
		pattern    string
		method     string
		rootRouter *Router
	}

	testCase := []TestCase{
		{
			name:    "user",
			pattern: "/user",
			method:  "GET",
			rootRouter: &Router{
				roots: map[string]*node{
					"GET": &node{
						part: "GET",
						children: []*node{
							{
								pattern: "/user",
								part:    "user",
							},
						},
					},
					"POST": &node{
						part: "POST",
						children: []*node{
							{
								pattern: "/user",
								part:    "user",
							},
						},
					},
				},
			},
		},
	}
	for _, tCase := range testCase {
		n, p := tCase.rootRouter.getRouter("GET", "/user")
		t.Logf("node:%v params:%v", n, p)
		// if n == nil {
		// 	t.Fatal()
		// }
	}
}
