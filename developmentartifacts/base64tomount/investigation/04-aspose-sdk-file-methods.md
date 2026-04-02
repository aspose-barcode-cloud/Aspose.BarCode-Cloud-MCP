# Aspose Barcode Cloud Go SDK: File-Based API Methods

## SDK Version

`github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4` v4.2603.0

## Currently Used Methods (Base64)

| Tool | SDK Method | Input |
|------|-----------|-------|
| `generate_barcode` | `GenerateAPI.Generate()` | params -> returns `[]byte` |
| `recognize_barcode` | `RecognizeAPI.RecognizeBase64()` | `RecognizeBase64Request{FileBase64: string}` |
| `scan_barcode` | `ScanAPI.ScanBase64()` | `ScanBase64Request{FileBase64: string}` |

## All Available API Methods

### GenerateAPI (3 methods)

```go
// Currently used - returns raw image bytes
Generate(ctx, barcodeType, data, *GenerateAPIGenerateOpts) ([]byte, *http.Response, error)

// Alternative with body params
GenerateBody(ctx, GenerateParams) ([]byte, *http.Response, error)

// Multipart form version
GenerateMultipart(ctx, barcodeType, data, *GenerateAPIGenerateMultipartOpts) ([]byte, *http.Response, error)
```

All generation methods return `[]byte`. None write directly to disk - the server must handle file writing.

### RecognizeAPI (3 methods)

```go
// URL-based - reads image from URL
Recognize(ctx, barcodeType, fileUrl string, *RecognizeAPIRecognizeOpts) (BarcodeResponseList, *http.Response, error)

// Currently used - accepts base64 string
RecognizeBase64(ctx, RecognizeBase64Request) (BarcodeResponseList, *http.Response, error)

// FILE-BASED - accepts *os.File handle
RecognizeMultipart(ctx, barcodeType, file *os.File, *RecognizeAPIRecognizeMultipartOpts) (BarcodeResponseList, *http.Response, error)
```

### ScanAPI (3 methods)

```go
// URL-based - reads image from URL
Scan(ctx, fileUrl string) (BarcodeResponseList, *http.Response, error)

// Currently used - accepts base64 string
ScanBase64(ctx, ScanBase64Request) (BarcodeResponseList, *http.Response, error)

// FILE-BASED - accepts *os.File handle
ScanMultipart(ctx, file *os.File) (BarcodeResponseList, *http.Response, error)
```

## Key Finding: File-Based Alternatives Exist

For recognize and scan operations, the SDK provides `*Multipart` methods that accept `*os.File`:

```go
// Instead of:
client.API.RecognizeAPI.RecognizeBase64(ctx, RecognizeBase64Request{FileBase64: base64String})

// Could use:
file, _ := os.Open("/mounted/path/to/image.png")
defer file.Close()
client.API.RecognizeAPI.RecognizeMultipart(ctx, barcodeType, file, opts)
```

```go
// Instead of:
client.API.ScanAPI.ScanBase64(ctx, ScanBase64Request{FileBase64: base64String})

// Could use:
file, _ := os.Open("/mounted/path/to/image.png")
defer file.Close()
client.API.ScanAPI.ScanMultipart(ctx, file)
```

## Notes

- `RecognizeMultipart` requires `barcodeType` as a separate parameter (not in the request body like `RecognizeBase64`)
- `ScanMultipart` has no barcode type parameter (auto-detect only)
- Internally, SDK reads the entire file with `io.ReadAll()` before sending as multipart form
- The methods accept `*os.File` specifically, not `io.Reader` interface
- For generation, `[]byte` must be written to disk manually with `os.WriteFile()`
