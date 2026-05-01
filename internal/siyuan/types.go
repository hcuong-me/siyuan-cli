// Package siyuan provides types for the SiYuan API.
package siyuan

// Notebook represents a SiYuan notebook.
type Notebook struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Sort   int    `json:"sort"`
	Closed bool   `json:"closed"`
}

// ListNotebooksResponse represents the response from /api/notebook/lsNotebooks.
type ListNotebooksResponse struct {
	Notebooks []Notebook `json:"notebooks"`
}

// CreateNotebookRequest represents the request for creating a notebook.
type CreateNotebookRequest struct {
	Name string `json:"name"`
}

// CreateNotebookResponse represents the response from creating a notebook.
type CreateNotebookResponse struct {
	Notebook Notebook `json:"notebook"`
}

// NotebookIDRequest represents a request that only needs notebook ID.
type NotebookIDRequest struct {
	Notebook string `json:"notebook"`
}

// RenameNotebookRequest represents the request for renaming a notebook.
type RenameNotebookRequest struct {
	Notebook string `json:"notebook"`
	Name     string `json:"name"`
}
