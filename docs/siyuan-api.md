# SiYuan API Reference

Complete API reference for SiYuan Note. Base URL: `http://127.0.0.1:6806`

## Authentication

All API requests require authentication via the `Authorization` header:

```
Authorization: Token {your-api-token}
```

Get your token from **Settings > About** in SiYuan.

## Response Format

All responses follow this structure:

```json
{
  "code": 0,
  "msg": "",
  "data": {}
}
```

- `code`: 0 for success, non-zero for errors
- `msg`: Error message (empty on success)
- `data`: Response data (varies by endpoint)

---

## System

### Get Version

```http
POST /api/system/version
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": "3.0.0"
}
```

### Get System Time

```http
POST /api/system/currentTime
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": 1743523200000
}
```

### Get Boot Progress

```http
POST /api/system/bootProgress
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "progress": 100,
    "stage": "Ready"
  }
}
```

---

## Notebooks

### List Notebooks

```http
POST /api/notebook/lsNotebooks
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "notebooks": [
      {
        "id": "20210817205410-2kvfpfn",
        "name": "Test Notebook",
        "icon": "1f41b",
        "sort": 0,
        "closed": false
      }
    ]
  }
}
```

### Open Notebook

```http
POST /api/notebook/openNotebook
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0"
}
```

### Close Notebook

```http
POST /api/notebook/closeNotebook
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0"
}
```

### Create Notebook

```http
POST /api/notebook/createNotebook
```

**Body:**
```json
{
  "name": "Notebook name"
}
```

### Rename Notebook

```http
POST /api/notebook/renameNotebook
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0",
  "name": "New name"
}
```

### Remove Notebook

```http
POST /api/notebook/removeNotebook
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0"
}
```

### Get Notebook Configuration

```http
POST /api/notebook/getNotebookConf
```

**Body:**
```json
{
  "notebook": "20210817205410-2kvfpfn"
}
```

### Set Notebook Configuration

```http
POST /api/notebook/setNotebookConf
```

**Body:**
```json
{
  "notebook": "20210817205410-2kvfpfn",
  "conf": {
    "name": "Test Notebook",
    "closed": false,
    "dailyNoteSavePath": "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}",
    "dailyNoteTemplatePath": ""
  }
}
```

---

## Documents

### List Document Tree

```http
POST /api/filetree/listDocTree
```

**Body:**
```json
{
  "notebook": "20210817205410-2kvfpfn",
  "path": "/"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "path": "/",
    "files": [
      {
        "path": "/doc1.sy",
        "name": "doc1",
        "id": "20200812220555-lj3enxa",
        "count": 5,
        "size": 1235,
        "updated": "20220428104712",
        "created": "20220428104712"
      }
    ]
  }
}
```

### Get Document IDs by Path

```http
POST /api/filetree/getIDsByHPath
```

**Body:**
```json
{
  "notebook": "20200812223209-lj3enxa",
  "path": "/Notes/Document"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": ["20200813125239-hbwpz87"]
}
```

### Create Document with Markdown

```http
POST /api/filetree/createDocWithMd
```

**Body:**
```json
{
  "notebook": "20210817205410-2kvfpfn",
  "path": "/foo/bar",
  "markdown": "# Title\n\nContent",
  "title": "Document Title"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "id": "20230601120000-abcdef1",
    "box": "20210817205410-2kvfpfn",
    "path": "/20230601120000-abcdef1.sy",
    "hPath": "/foo/bar/Document Title"
  }
}
```

### Rename Document

```http
POST /api/filetree/renameDoc
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0",
  "path": "/20210902210113-0avi12f.sy",
  "title": "New title"
}
```

Or by ID:

```http
POST /api/filetree/renameDocByID
```

**Body:**
```json
{
  "id": "20210902210113-0avi12f",
  "title": "New title"
}
```

### Remove Document

```http
POST /api/filetree/removeDoc
```

**Body:**
```json
{
  "notebook": "20210831090520-7dvbdv0",
  "path": "/20210902210113-0avi12f.sy"
}
```

Or by ID:

```http
POST /api/filetree/removeDocByID
```

**Body:**
```json
{
  "id": "20210902210113-0avi12f"
}
```

### Move Documents

```http
POST /api/filetree/moveDocs
```

**Body:**
```json
{
  "fromPaths": ["/20210917220056-yxtyl7i.sy"],
  "toNotebook": "20210817205410-2kvfpfn",
  "toPath": "/"
}
```

---

## Blocks

### Get Block Kramdown

```http
POST /api/block/getBlockKramdown
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

### Get Child Blocks

```http
POST /api/block/getChildBlocks
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

### Insert Block

```http
POST /api/block/insertBlock
```

**Body:**
```json
{
  "dataType": "markdown",
  "data": "Content",
  "parentID": "20200812220555-lj3enxa"
}
```

### Update Block

```http
POST /api/block/updateBlock
```

**Body:**
```json
{
  "dataType": "markdown",
  "data": "Updated content",
  "id": "20200812220555-lj3enxa"
}
```

### Delete Block

```http
POST /api/block/deleteBlock
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

---

## Attributes

### Get Block Attributes

```http
POST /api/attr/getBlockAttrs
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "custom-key": "value"
  }
}
```

### Set Block Attributes

```http
POST /api/attr/setBlockAttrs
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa",
  "attrs": {
    "custom-key": "value"
  }
}
```

---

## SQL Query

### Execute SQL

```http
POST /api/query/sql
```

**Body:**
```json
{
  "stmt": "SELECT * FROM blocks LIMIT 10"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": [
    {
      "id": "20200812220555-lj3enxa",
      "content": "Block content",
      "type": "p"
    }
  ]
}
```

**Security:** Only SELECT queries are allowed.

---

## Search

### Full-Text Search

```http
POST /api/search/fullTextSearchBlock
```

**Body:**
```json
{
  "k": "keyword",
  "page": 1,
  "size": 20
}
```

### Search Documents

```http
POST /api/filetree/searchDocs
```

**Body:**
```json
{
  "k": "keyword",
  "box": "notebook-id"
}
```

---

## Tags

### Get All Tags

```http
POST /api/tag/getTag
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": [
    {
      "label": "tag-name",
      "count": 5
    }
  ]
}
```

### Search Tags

```http
POST /api/search/searchTag
```

**Body:**
```json
{
  "k": "keyword"
}
```

---

## Templates

### List Templates

```http
POST /api/file/readDir
```

**Body:**
```json
{
  "path": "/data/templates"
}
```

### Get Template

```http
POST /api/file/getFile
```

**Body:**
```json
{
  "path": "/data/templates/template.md"
}
```

### Render Template

```http
POST /api/template/render
```

**Body:**
```json
{
  "id": "20220724223548-j6g0o87",
  "path": "/data/templates/foo.md"
}
```

---

## Assets

### Upload Asset

```http
POST /api/asset/upload
Content-Type: multipart/form-data
```

**Form Fields:**
- `assetsDirPath`: `/assets/`
- `file[]`: File content

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "succMap": {
      "foo.png": "assets/foo-20210719092549-9j5y79r.png"
    },
    "errFiles": []
  }
}
```

### Get Unused Assets

```http
POST /api/asset/getUnusedAssets
```

### Remove Unused Assets

```http
POST /api/asset/removeUnusedAssets
```

---

## File Operations

### Read Directory

```http
POST /api/file/readDir
```

**Body:**
```json
{
  "path": "/data/assets"
}
```

### Get File

```http
POST /api/file/getFile
```

**Body:**
```json
{
  "path": "/data/storage/file.md"
}
```

### Put File

```http
POST /api/file/putFile
```

**Body:**
```json
{
  "path": "/data/storage/file.md",
  "content": "File content",
  "isDir": false
}
```

### Remove File

```http
POST /api/file/removeFile
```

**Body:**
```json
{
  "path": "/data/storage/file.md"
}
```

### Rename File

```http
POST /api/file/renameFile
```

**Body:**
```json
{
  "path": "/old/path",
  "newPath": "/new/path"
}
```

---

## Export

### Export Markdown

```http
POST /api/export/exportMdContent
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "content": "# Markdown content"
  }
}
```

### Export HTML

```http
POST /api/export/exportHTML
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa"
}
```

### Export PDF

```http
POST /api/export/exportPDF
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa",
  "savePath": "/tmp",
  "removeAssets": false
}
```

### Export DOCX

```http
POST /api/export/exportDocx
```

**Body:**
```json
{
  "id": "20200812220555-lj3enxa",
  "savePath": "/tmp",
  "removeAssets": false
}
```

---

## Notifications

### Push Message

```http
POST /api/notification/pushMsg
```

**Body:**
```json
{
  "msg": "Hello World",
  "timeout": 7000
}
```

### Push Error Message

```http
POST /api/notification/pushErrMsg
```

**Body:**
```json
{
  "msg": "Error occurred"
}
```

---

## Snapshots (Repository)

### List Snapshots

```http
POST /api/repo/getRepoSnapshots
```

### Get Current Snapshot

```http
POST /api/repo/getRepoSnapshot
```

### Create Snapshot

```http
POST /api/repo/createRepoSnapshot
```

**Body:**
```json
{
  "memo": "Backup before update"
}
```

### Restore Snapshot

```http
POST /api/repo/checkoutRepo
```

**Body:**
```json
{
  "id": "snapshot-id"
}
```

### Remove Snapshot

```http
POST /api/repo/removeRepoSnapshot
```

**Body:**
```json
{
  "id": "snapshot-id"
}
```
