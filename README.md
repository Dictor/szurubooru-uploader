# szurubooru-uploader
forder uploader with tag assigning feature for szurubooru

## Usage
- `dir` is necessary option.
- if didn't give `uid` or `upw` or both, program ask these interactive
- currently `tag` option support only one tag. (if you can, please contribute!)

```
./szurubooru-uploader -<option1> "<value>" -<option2> "<value2>"...
  -debug
        print debug log
  -dir string
        directory to upload
  -host string
        address of host (default "http://localhost")
  -safety string
        safety of images in directory (default "unsafe")
  -tag string
        tag which will be assigned to images
  -uid string
        user's login id
  -upw string
        user's login password
```
