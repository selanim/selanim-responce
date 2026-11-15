# selanim-responce

`selanim-responce` ni package nyepesi ya Go ya kuandika HTTP JSON responses
zenye muundo unaofanana. Inajumuisha success, error codes, validation error,
RFC 9457 Problem Details, no-content responses, page/cursor pagination,
configurable logging na request ID middleware. Haina dependency za nje.

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

Error hujibu kwa `success: false` na `message`; inaweza pia kuwa na `code`.
Validation error hujumuisha `errors`, na pagination hujumuisha `meta`.

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

## Logging

Package hutumia logger ya kawaida ya Go kuandika status code baada ya response
kuandikwa, na kuandika makosa ya status code isiyo sahihi, JSON encoding,
kuandika response au pagination. Body ya response haiwekwi kwenye log. Errors
bado hurudishwa kwa caller, kwa hiyo unaweza kuzishughulikia au kuongeza
context kwenye logs zako. Kwa default hutumia `log.Default()`. Weka logger yako
kwa `SetLogger`; pitisha `nil` kuzima logging ya package:

```go
responce.SetLogger(log.New(os.Stderr, "api: ", log.LstdFlags))
```

Logger unayoweka inapaswa kuwa salama kutumiwa na goroutines nyingi.

## Request ID

`RequestIDMiddleware` hutengeneza ID mpya ya kipekee kwa kila request na
huiweka kwenye response header `X-Request-ID` na request context. Thamani
iliyotumwa kwenye header hiyo na client hubadilishwa. Handler inaweza kupata ID
kwa `RequestIDFromContext`:

```go
handler := responce.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	requestID := responce.RequestIDFromContext(r.Context())
	_ = responce.WriteSuccess(w, http.StatusOK, "Imekamilika", map[string]string{
		"request_id": requestID,
	})
}))
```

## Success response

`WriteSuccess` inachukua `http.ResponseWriter`, HTTP status code, ujumbe na data
yoyote inayoweza ku-encode kuwa JSON:

```go
err := responce.WriteSuccess(w, http.StatusCreated, "User ameundwa", user)
if err != nil {
	log.Printf("andika response: %v", err)
}
```

## No content response

Tumia `WriteNoContent` kwa operesheni iliyofanikiwa ambayo haina body ya kurudisha,
kama kufuta rasilimali. Inaandika HTTP `204` bila JSON body:

```go
if err := responce.WriteNoContent(w); err != nil {
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

Ikiwa client anahitaji kutambua kosa kwa code isiyotegemea maandishi, tumia
`WriteErrorWithCode`. Code huongezwa kwenye JSON kama `code`:

```go
err := responce.WriteErrorWithCode(
	w,
	http.StatusNotFound,
	"USER_NOT_FOUND",
	"User hapatikani",
)
```

Kwa validation errors zenye code, tumia `WriteValidationErrorWithCode`.

## RFC 9457 Problem Details

Ikiwa API inahitaji error format inayotumika kwa pamoja na HTTP APIs nyingine,
tumia `WriteProblem`. HTTP status na `status` ya JSON huwekwa zilingane; `Type`
isipojazwa, hutumia `about:blank`. `Extensions` huongeza fields maalum kwenye
kiwango cha juu cha JSON:

```go
err := responce.WriteProblem(w, http.StatusNotFound, responce.ProblemDetails{
	Type:   "https://example.com/problems/user-not-found",
	Title:  "User hapatikani",
	Detail: "Hakuna user mwenye ID hiyo.",
	Extensions: map[string]any{
		"code": "USER_NOT_FOUND",
	},
})
```

Response hutumia `Content-Type: application/problem+json`. Tumia `WriteError`
au `WriteErrorWithCode` ikiwa unataka envelope ya kawaida ya package badala yake.

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

## Headers za ziada

`WriteJSONWithHeaders` inaruhusu kuongeza headers kwa JSON response. Thamani za
headers zinazotolewa hubadilisha zilizopo zenye jina sawa; `Content-Type`
hubaki kuwa JSON:

```go
err := responce.WriteJSONWithHeaders(
	w,
	http.StatusOK,
	responce.Response{Success: true, Message: "Imekamilika"},
	http.Header{"Cache-Control": {"no-store"}},
)
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

### Link headers

`WritePaginatedWithLinks` huandika response ya pagination pamoja na RFC 8288
`Link` header ya `first`, `prev`, `next` na `last` pale inapofaa. Inahifadhi
query parameters nyingine za request URL, na hubadilisha `page` na `per_page`
kwenye links.

```go
err := responce.WritePaginatedWithLinks(
	w,
	r,
	http.StatusOK,
	"Users wamepatikana",
	users,
	page,
	perPage,
	totalItems,
)
```

### Cursor pagination

Kwa data inayobadilika mara kwa mara au inayohitaji kusogeza ukurasa kwa ufanisi,
tumia `WriteCursorPaginated`. Package huhifadhi cursors kama tokens zisizo na
maana kwake; application yako ndiyo inayotengeneza na kutafsiri cursors thabiti
kulingana na sorting ya database:

```go
err := responce.WriteCursorPaginated(
	w,
	http.StatusOK,
	"Users wamepatikana",
	users,
	nextCursor,
	previousCursor,
)
```

Cursor isiyo tupu huashiria ukurasa unaofuata au uliotangulia na huwekwa kwenye
`meta.next_cursor` au `meta.previous_cursor`.

Mifano ya `RequestIDMiddleware` na `WritePaginated` iko katika Go examples
zinazoendeshwa pamoja na tests kwa `go test ./...`.

## API

| Function/Type | Maelezo |
| --- | --- |
| `WriteJSON(w, statusCode, payload)` | Encode payload kama JSON na kuandika HTTP response. |
| `WriteJSONWithHeaders(w, statusCode, payload, headers)` | Kuandika JSON response yenye headers za ziada. |
| `WriteNoContent(w)` | Kuandika HTTP 204 bila response body. |
| `WriteSuccess(w, statusCode, message, data)` | Kuandika success response. |
| `WriteError(w, statusCode, message)` | Kuandika error response. |
| `WriteErrorWithCode(w, statusCode, code, message)` | Kuandika error response yenye code. |
| `WriteValidationError(w, message, errors)` | Kuandika validation errors kwa HTTP 422. |
| `WriteValidationErrorWithCode(w, code, message, errors)` | Kuandika validation errors zenye code kwa HTTP 422. |
| `WriteProblem(w, statusCode, problem)` | Kuandika RFC 9457 Problem Details response. |
| `WritePaginated(w, statusCode, message, data, page, perPage, totalItems)` | Kuandika success response yenye pagination metadata. |
| `WritePaginatedWithLinks(w, r, statusCode, message, data, page, perPage, totalItems)` | Kuandika pagination response yenye `Link` headers. |
| `WriteCursorPaginated(w, statusCode, message, data, nextCursor, previousCursor)` | Kuandika pagination response yenye cursors. |
| `SetLogger(logger)` | Kusanidi logger ya package; `nil` huzima logging. |
| `RequestIDMiddleware(next)` | Kutengeneza ID na kuiweka kwenye header na request context. |
| `RequestIDFromContext(ctx)` | Kusoma request ID iliyowekwa na middleware. |
| `NewPaginationMeta(page, perPage, totalItems)` | Kuhesabu na kuthibitisha pagination metadata. |
| `PaginationMeta.Offset()` | Kuhesabu offset ya ukurasa kwa database query. |
| `PaginationMeta` / `CursorPaginationMeta` | Metadata ya page au cursor pagination. |
| `ProblemDetails` | Muundo wa RFC 9457 Problem Details. |
| `Logger` | Interface ya logger inayopokea `Printf`. |
| `Response` | Muundo wa kawaida wa JSON response. |
| `ValidationErrors` | Map ya field kwenda kwenye orodha ya ujumbe wa validation. |
