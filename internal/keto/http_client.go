package keto

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CheckRequest struct {
    Namespace string `json:"namespace"`
    Object    string `json:"object"`
    Relation  string `json:"relation"`
    SubjectID string `json:"subject_id"`
}

type CheckResponse struct {
    Allowed bool `json:"allowed"`
}

func CheckPermission(namespace, object, relation, subjectID string) bool {
    reqBody := CheckRequest{
        Namespace: namespace,
        Object:    object,
        Relation:  relation,
        SubjectID: subjectID,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        fmt.Printf("Error marshaling request: %v\n", err)
        return false
    }

    resp, err := http.Post("http://localhost:4466/relation-tuples/check", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("HTTP request failed: %v\n", err)
        return false
    }
    defer resp.Body.Close()

    var checkResp CheckResponse
    if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
        fmt.Printf("Error decoding response: %v\n", err)
        return false
    }

    return checkResp.Allowed
}
