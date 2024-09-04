package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type RSS struct {
	XMLName xml.Name  `xml:"root"`
	Users   []UserRow `xml:"row"`
}

type UserRow struct {
	Id        int `xml:"id"`
	Name      string
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

const filePath = "dataset.xml"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	tkn := r.Header.Get("AccessToken")
	if tkn == "" || tkn != os.Getenv("ACCESS_TOKEN") {
		http.Error(w, "bad Access Token ", http.StatusUnauthorized)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
		//return fmt.Errorf("not open file: %w", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var rss RSS
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, user := range rss.Users {
		fullName := user.FirstName + " " + user.LastName
		rss.Users[i].Name = fullName
	}

	offset := r.URL.Query().Get("offset")
	if offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if offsetInt >= len(rss.Users) {
			http.Error(w, "offset is out of range for the list of users", http.StatusInternalServerError)
			return
		}

		rss.Users = rss.Users[offsetInt:]
	}

	query := r.URL.Query().Get("query")
	if query != "" {
		switch {
		case strings.HasPrefix(query, "about="):
			rss.Users = filterUsersByAbout(rss.Users, strings.TrimPrefix(query, "about="))
		case strings.HasPrefix(query, "name="):
			rss.Users = filterUsersByName(rss.Users, strings.TrimPrefix(query, "name="))
		}
	}

	orderField := r.URL.Query().Get("order_field")
	orderBy := r.URL.Query().Get("order_by")
	if orderField != "" && orderBy != "" {
		if orderField == "Id" || orderField == "Name" || orderField == "Age" {
			rss.Users = sortUsers(rss.Users, orderBy, orderField)
		} else {
			errResp, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
			http.Error(w, string(errResp), http.StatusBadRequest)
			return
		}
	} else if orderBy != "" {
		rss.Users = sortUsers(rss.Users, orderBy, "Name")
	}

	limit := r.URL.Query().Get("limit")
	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			errResp, _ := json.Marshal(SearchErrorResponse{Error: "Limit convert to int"})
			http.Error(w, string(errResp), http.StatusBadRequest)
			return
		}

		if limitInt > len(rss.Users) {
			limitInt = len(rss.Users)
		}

		rss.Users = rss.Users[:limitInt]
	}

	res, _ := json.Marshal(rss.Users)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func filterUsersByName(users []UserRow, query string) []UserRow {
	filteredUsers := make([]UserRow, 0)
	for _, user := range users {
		if strings.Contains(user.Name, query) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func filterUsersByAbout(users []UserRow, query string) []UserRow {
	filteredUsers := make([]UserRow, 0)
	for _, user := range users {
		if strings.Contains(user.About, query) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func sortUsers(users []UserRow, sortType string, sortField string) []UserRow {
	sort.Slice(users, func(i, j int) bool {
		switch sortField {
		case "Name":
			if sortType == strconv.Itoa(OrderByDesc) {
				return users[i].Name > users[j].Name
			} else if sortType == strconv.Itoa(OrderByAsc) {
				return users[i].Name < users[j].Name
			} else {
				return false
			}
		case "Age":
			if sortType == strconv.Itoa(OrderByDesc) {
				return users[i].Age > users[j].Age
			} else {
				return users[i].Age < users[j].Age
			}
		case "Id":
			if sortType == strconv.Itoa(OrderByDesc) {
				return users[i].Id > users[j].Id
			} else {
				return users[i].Id < users[j].Id
			}
		default:
			return false
		}
	})
	return users
}
