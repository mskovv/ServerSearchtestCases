package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

type TestCase struct {
	Name            string
	srv             SearchClient
	req             SearchRequest
	want            *SearchResponse
	wantErr         bool
	wantErrContains string
}

func checkWantErr(t *testing.T, got *SearchResponse, err error, testCase TestCase) {
	if err != nil {
		if !testCase.wantErr {
			t.Errorf("FindUsers() unexpected error = %v", err)
			return
		}
		if testCase.wantErrContains != "" && !strings.Contains(err.Error(), testCase.wantErrContains) {
			t.Errorf("FindUsers() got error = %v, want error containing %q", err, testCase.wantErrContains)
			return
		}
	} else {
		if testCase.wantErr {
			t.Errorf("FindUsers() expected error, but got none")
			return
		}
	}
	if !reflect.DeepEqual(got, testCase.want) {
		t.Errorf("FindUsers() got = %v, want %v", got, testCase.want)
	}
}

func TestSearchClient_FindUsers(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	srv := &SearchClient{
		AccessToken: os.Getenv("ACCESS_TOKEN"),
		URL:         ts.URL,
	}

	cases := []TestCase{
		{
			Name: "Good One User",
			srv:  *srv,
			req: SearchRequest{
				Limit:   1,
				Offset:  1,
				OrderBy: OrderByAsIs,
			},
			want: &SearchResponse{
				Users: []User{
					{
						Id:     1,
						Name:   "Hilda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			wantErr: false,
		},
		{
			Name: "Bad limit",
			srv:  *srv,
			req: SearchRequest{
				Limit:   -1,
				Offset:  1,
				OrderBy: OrderByAsIs,
			},
			want:    nil,
			wantErr: true,
		},
		{
			Name: "Bad limit",
			srv:  *srv,
			req: SearchRequest{
				Limit:   26, // > 25 ? limit = 25
				Offset:  34,
				OrderBy: OrderByAsIs,
			},
			want: &SearchResponse{
				Users: []User{
					{
						Id:     34,
						Name:   "Kane Sharp",
						Age:    34,
						About:  "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			wantErr: false,
		},
		{
			Name: "Bad access token",
			srv: SearchClient{
				AccessToken: "123312",
				URL:         ts.URL,
			},
			req:     SearchRequest{},
			want:    nil,
			wantErr: true,
		},
		{
			Name: "Bad offset",
			srv:  *srv,
			req: SearchRequest{
				Limit:  5,
				Offset: -1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			Name: "StatusInternalServerError",
			srv:  *srv,
			req: SearchRequest{
				Offset: 40, // offset > 35 = StatusInternalServerError
			},
			want:    nil,
			wantErr: true,
		},
		{
			Name: "StatusBadRequest error bad order field",
			srv:  *srv,
			req: SearchRequest{
				OrderField: "Biba",
			},
			want:            nil,
			wantErr:         true,
			wantErrContains: "OrderFeld Biba invalid",
		},
		{
			Name: "Query by about",
			srv:  *srv,
			req: SearchRequest{
				Limit: 25,
				Query: "about=Nulla cillum enim voluptate consequat laborum",
			},
			want: &SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			wantErr: false,
		},
		{
			Name: "Query by Name",
			srv:  *srv,
			req: SearchRequest{
				Limit: 25,
				Query: "name=Boyd Wolf",
			},
			want: &SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			wantErr: false,
		},
		{
			Name: "StatusInternalServerError",
			srv:  *srv,
			req: SearchRequest{
				Offset: 40, // offset > 35 = StatusInternalServerError
			},
			want:    nil,
			wantErr: true,
		},
		{
			Name: "SortByOrderField",
			srv:  *srv,
			req: SearchRequest{
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
				Offset:     33,
				Limit:      1,
			},
			want: &SearchResponse{
				Users: []User{
					{
						Id:     33,
						Name:   "Twila Snow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := tc.srv.FindUsers(tc.req)
			checkWantErr(t, got, err, tc)
		})
	}

	tsWithTimeOut := httptest.NewServer(http.TimeoutHandler(http.HandlerFunc(SearchServer), time.Second, "Timeout"))
	defer tsWithTimeOut.Close()

	t.Run("timeoutCase with timer", func(t *testing.T) {
		testCase := TestCase{
			Name: "9 Timeout",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsWithTimeOut.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "timeout",
		}

		resCh := make(chan *SearchResponse, 1)
		errCh := make(chan error, 1)
		go func() {
			res, err := testCase.srv.FindUsers(testCase.req)
			if err != nil {
				errCh <- err
			} else {
				resCh <- res
			}
		}()

		select {
		case <-time.After(time.Second * 2):
			// Таймаут
		case res := <-resCh:
			if res == nil {
				// Ответ не должен быть nil
				t.Errorf("FindUsers() returned nil response")
				return
			}
		case err := <-errCh:
			if !testCase.wantErr {
				t.Errorf("FindUsers() unexpected error = %v", err)
				return
			}
			if !strings.Contains(err.Error(), testCase.wantErrContains) {
				t.Errorf("FindUsers() got error = %v, want error containing %q", err, testCase.wantErrContains)
				return
			}
		}

	})

	tsWithError := httptest.NewUnstartedServer(http.HandlerFunc(SearchServer))

	t.Run("case with server start error", func(t *testing.T) {
		testCase := TestCase{
			Name: "Error not start",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsWithError.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "error",
		}

		got, err := testCase.srv.FindUsers(testCase.req)
		checkWantErr(t, got, err, testCase)
	})

	tsWithTimeoutError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Ok"))
		if err != nil {
			return
		}
	}))

	t.Run("case with timeout error", func(t *testing.T) {
		testCase := TestCase{
			Name: "Error not start",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsWithTimeoutError.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "timeout",
		}

		got, err := testCase.srv.FindUsers(testCase.req)
		checkWantErr(t, got, err, testCase)
	})

	tsWithInvalidJson := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid json", http.StatusBadRequest)
	}))

	t.Run("case with Invalid json", func(t *testing.T) {
		testCase := TestCase{
			Name: "Error not start",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsWithInvalidJson.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "cant unpack error json: invalid character",
		}

		got, err := testCase.srv.FindUsers(testCase.req)
		checkWantErr(t, got, err, testCase)
	})

	tsWithValidJsonBadRequest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errResp, _ := json.Marshal(SearchErrorResponse{Error: "Valid"})
		http.Error(w, string(errResp), http.StatusBadRequest)
	}))

	t.Run("case with valid json in bad request", func(t *testing.T) {
		testCase := TestCase{
			Name: "Error not start",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsWithValidJsonBadRequest.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "unknown bad request error:",
		}

		got, err := testCase.srv.FindUsers(testCase.req)
		checkWantErr(t, got, err, testCase)
	})

	tsBadBodyToUnmarshal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, _ := json.Marshal(" ")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(res)

	}))

	t.Run("case with valid json in bad request", func(t *testing.T) {
		testCase := TestCase{
			Name: "Error not start",
			srv: SearchClient{
				AccessToken: srv.AccessToken,
				URL:         tsBadBodyToUnmarshal.URL,
			},
			req:             SearchRequest{},
			want:            nil,
			wantErr:         true,
			wantErrContains: "cant unpack result json",
		}

		got, err := testCase.srv.FindUsers(testCase.req)
		checkWantErr(t, got, err, testCase)
	})

}
