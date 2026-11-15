# selanim-responce

`selanim-responce` ni package nyepesi ya Go ya kuandika HTTP JSON responses
zenye muundo unaofanana. Inajumuisha success, error, validation error na
pagination metadata. Haina dependency za nje.

## Mahitaji

- Go 1.21 au toleo jipya zaidi.
- Module hii inapatikana GitHub kwa `github.com/selanim/selanim-responce`.

## Kusakinisha

Ndani ya project yako ya Go, endesha:

```sh
go get github.com/selanim/selanim-responce
```

Kisha import package. Jina lake ndani ya Go ni `responce`:

```go
import "github.com/selanim/selanim-responce"
```

## Muundo wa response

Responses hutumia envelope hii. Fields zisizo na thamani zinaweza kuachwa:

```json
{
  "success": true,
  "message": "User amepatikana",
  "data": {
    "id": 1,
    "name": "Amina"
  }
}
```

Error hujibu kwa `success: false` na `message`; validation error pia hujumuisha
`errors`. Pagination hujumuisha `meta`.

## Mfano kamili wa HTTP handler

```go
package main

import (
	"log"
	"net/http"

	responce "github.com/selanim/selanim-responce"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "Amina"},
		{ID: 2, Name: "Juma"},
	}

	if err := responce.WriteSuccess(w, http.StatusOK, "Users wamepatikana", users); err != nil {
		log.Printf("andika response: %v", err)
	}
}

func main() {
	http.HandleFunc("/users", usersHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Kwenye handler zako, tumia `http.StatusOK`, `http.StatusCreated`,
`http.StatusNotFound` na HTTP status nyingine inayofaa. Kagua `error` inayorudishwa
na helper; ikiwa response tayari imeanza kuandikwa, usijaribu kuandika response
nyingine.

## Success response

`WriteSuccess` inachukua `http.ResponseWriter`, HTTP status code, ujumbe na data
yoyote inayoweza ku-encode kuwa JSON:

```go
err := responce.WriteSuccess(w, http.StatusCreated, "User ameundwa", user)
if err != nil {
	log.Printf("andika response: %v", err)
}
```

## Error response

Tumia `WriteError` kwa errors za kawaida. Wewe ndiye unachagua HTTP status code:

```go
err := responce.WriteError(w, http.StatusNotFound, "User hapatikani")
if err != nil {
	log.Printf("andika response: %v", err)
}
```

## Validation errors

`WriteValidationError` huandika HTTP `422 Unprocessable Entity`. Tuma field
inayokosewa na orodha ya sababu zake:

```go
err := responce.WriteValidationError(
	w,
	"Taarifa ulizoweka si sahihi",
	responce.ValidationErrors{
		"email":    {"email inahitajika", "lazima iwe email sahihi"},
		"password": {"lazima iwe na angalau herufi 8"},
	},
)
if err != nil {
	log.Printf("andika response: %v", err)
}
```

Mfano wa JSON:

```json
{
  "success": false,
  "message": "Taarifa ulizoweka si sahihi",
  "errors": {
    "email": ["email inahitajika", "lazima iwe email sahihi"],
    "password": ["lazima iwe na angalau herufi 8"]
  }
}
```

## Pagination

`WritePaginated` huandika data ya ukurasa pamoja na metadata. `page` huanza
kwenye `1`, `perPage` ni idadi ya rekodi kwa ukurasa, na `totalItems` ni jumla
ya rekodi zote zinazolingana na query (si idadi ya ukurasa wa sasa).

```go
err := responce.WritePaginated(
	w,
	http.StatusOK,
	"Users wamepatikana",
	users,       // data ya ukurasa wa sasa
	page,        // kwa mfano 2
	perPage,     // kwa mfano 10
	totalItems,  // jumla ya rekodi zote, kwa mfano 25
)
if err != nil {
	log.Printf("andika response: %v", err)
}
```

Matokeo ya mfano yana `total_pages: 3`, `has_next: true` na `has_previous: true`:

```json
{
  "success": true,
  "message": "Users wamepatikana",
  "data": [{"id": 11, "name": "Neema"}],
  "meta": {
    "page": 2,
    "per_page": 10,
    "total_items": 25,
    "total_pages": 3,
    "has_next": true,
    "has_previous": true
  }
}
```

### Kuhesabu page na database offset

Unaweza kupata `page` na `perPage` kutoka query parameters. Hakikisha unazikagua
na kuzikataa ikiwa si namba sahihi au ziko chini ya `1`:

```go
page, err := strconv.Atoi(r.URL.Query().Get("page"))
if err != nil || page < 1 {
	_ = responce.WriteError(w, http.StatusBadRequest, "page lazima iwe namba kuanzia 1")
	return
}

perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
if err != nil || perPage < 1 {
	_ = responce.WriteError(w, http.StatusBadRequest, "per_page lazima iwe namba kuanzia 1")
	return
}

meta, err := responce.NewPaginationMeta(page, perPage, totalItems)
if err != nil {
	_ = responce.WriteError(w, http.StatusBadRequest, "pagination parameters si sahihi")
	return
}
offset, err := meta.Offset()
if err != nil {
	log.Printf("hesabu pagination offset: %v", err)
	return
}
```

Tumia `perPage` na `offset` kwenye query ya database yako, kisha peleka rekodi
zilizopatikana pamoja na `totalItems` kwa `WritePaginated`. Database nyingi
zinahitaji query tofauti ya kuhesabu jumla ya rekodi na query ya kuchukua ukurasa.

`NewPaginationMeta(page, perPage, totalItems)` pia inapatikana bila kuandika
response. Inarudisha error ikiwa ukurasa/ukubwa wa ukurasa si halali au jumla ni
hasi. `PaginationMeta.Offset()` inarudisha offset ya kuanzia `0` na error ikiwa
hesabu imezidi ukubwa wa `int64`.

## API

| Function/Type | Maelezo |
| --- | --- |
| `WriteJSON(w, statusCode, payload)` | Encode payload kama JSON na kuandika HTTP response. |
| `WriteSuccess(w, statusCode, message, data)` | Kuandika success response. |
| `WriteError(w, statusCode, message)` | Kuandika error response. |
| `WriteValidationError(w, message, errors)` | Kuandika validation errors kwa HTTP 422. |
| `WritePaginated(w, statusCode, message, data, page, perPage, totalItems)` | Kuandika success response yenye pagination metadata. |
| `NewPaginationMeta(page, perPage, totalItems)` | Kuhesabu na kuthibitisha pagination metadata. |
| `PaginationMeta.Offset()` | Kuhesabu offset ya ukurasa kwa database query. |
| `Response` | Muundo wa kawaida wa JSON response. |
| `ValidationErrors` | Map ya field kwenda kwenye orodha ya ujumbe wa validation. |
