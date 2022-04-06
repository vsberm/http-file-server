package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	tarGzKey         = "tar.gz"
	tarGzValue       = "true"
	tarGzContentType = "application/x-tar+gzip"

	zipKey         = "zip"
	zipValue       = "true"
	zipContentType = "application/zip"

	osPathSeparator = string(filepath.Separator)
)

const directoryListingTemplateText = `
<!DOCTYPE html>
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <link rel="icon" type="image/png" sizes="16x16" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAd5JREFUeNqMU79rFUEQ/vbuodFEEkzAImBpkUabFP4ldpaJhZXYm/RiZWsv/hkWFglBUyTIgyAIIfgIRjHv3r39MePM7N3LcbxAFvZ2b2bn22/mm3XMjF+HL3YW7q28YSIw8mBKoBihhhgCsoORot9d3/ywg3YowMXwNde/PzGnk2vn6PitrT+/PGeNaecg4+qNY3D43vy16A5wDDd4Aqg/ngmrjl/GoN0U5V1QquHQG3q+TPDVhVwyBffcmQGJmSVfyZk7R3SngI4JKfwDJ2+05zIg8gbiereTZRHhJ5KCMOwDFLjhoBTn2g0ghagfKeIYJDPFyibJVBtTREwq60SpYvh5++PpwatHsxSm9QRLSQpEVSd7/TYJUb49TX7gztpjjEffnoVw66+Ytovs14Yp7HaKmUXeX9rKUoMoLNW3srqI5fWn8JejrVkK0QcrkFLOgS39yoKUQe292WJ1guUHG8K2o8K00oO1BTvXoW4yasclUTgZYJY9aFNfAThX5CZRmczAV52oAPoupHhWRIUUAOoyUIlYVaAa/VbLbyiZUiyFbjQFNwiZQSGl4IDy9sO5Wrty0QLKhdZPxmgGcDo8ejn+c/6eiK9poz15Kw7Dr/vN/z6W7q++091/AQYA5mZ8GYJ9K0AAAAAASUVORK5CYII=" />
    <title>{{ .Title }}</title>
    <style>
        h1 {
            border-bottom: 1px solid #c0c0c0;
            margin-bottom: 10px;
            padding-bottom: 10px;
            white-space: nowrap;
        }

        table {
            /* border-collapse: collapse; */
            /* border: 1px solid #c0c0c0; */
            width: 80%;
        }

        tr.header {
            font-weight: bold;
        }

        td.detailsColumn {
            -webkit-padding-start: 2em;
            padding-inline-start: 2em;
            text-align: end;
            white-space: nowrap;
        }

        a.icon {
            -webkit-padding-start: 1.5em;
            padding-inline-start: 1.5em;
            text-decoration: none;
            padding-left: 20px;
        }

        a.icon:hover {
            text-decoration: underline;
        }

        a.file {
            /* background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAIAAACQkWg2AAAABnRSTlMAAAAAAABupgeRAAABHUlEQVR42o2RMW7DIBiF3498iHRJD5JKHurL+CRVBp+i2T16tTynF2gO0KSb5ZrBBl4HHDBuK/WXACH4eO9/CAAAbdvijzLGNE1TVZXfZuHg6XCAQESAZXbOKaXO57eiKG6ft9PrKQIkCQqFoIiQFBGlFIB5nvM8t9aOX2Nd18oDzjnPgCDpn/BH4zh2XZdlWVmWiUK4IgCBoFMUz9eP6zRN75cLgEQhcmTQIbl72O0f9865qLAAsURAAgKBJKEtgLXWvyjLuFsThCSstb8rBCaAQhDYWgIZ7myM+TUBjDHrHlZcbMYYk34cN0YSLcgS+wL0fe9TXDMbY33fR2AYBvyQ8L0Gk8MwREBrTfKe4TpTzwhArXWi8HI84h/1DfwI5mhxJamFAAAAAElFTkSuQmCC ") left top no-repeat; */
            background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAUCAAAAAChCeKrAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QA/4ePzL8AAAAHdElNRQfmBAIGEyl0kJk4AAABOUlEQVQY0x2QzU7CQBSF5/F8CuMbuEEKtBDaIgKJMSa6MNHQosaFifyqcQOlGFmKC5XExAS1FKKmdMpP22lnvPXLmc3NPefcDErGMmkhDcqk1/Ypoyhm0YAQUMj2Ni8pQ5zTK5aAwi07GBxdMRTH2JyY5mRsseNye+MEBu+drqZp7SGbXqiH6yjhDNSz00pF7VPG2JJHHA494ns+BLvLwEmhuNMR5Zwsi9cfpe2XBWzYng3M7FWAMcEpqH2t1ur1WnXw02oYc7DgYaMJ1J5+b5rmXICBC/vgWQW2TRweaiFUkuRsa1TIPS8iS+hHtT6hxAujWvxYVhRFLfen58oIQjl71NV7PV1/mz3cf0e1GH8ZBshyzfHy/1I9vwPk7z53i8OFgLZm8EEBEIYkoDiJeE4Us9lI8KQE/wci5vKM6kYrRgAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyMi0wNC0wMlQwNjoxOToyNSswMDowMMFnhP4AAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjItMDQtMDJUMDY6MTk6MjUrMDA6MDCwOjxCAAAAAElFTkSuQmCC") left top no-repeat;
        }

        a.dir {
            background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAd5JREFUeNqMU79rFUEQ/vbuodFEEkzAImBpkUabFP4ldpaJhZXYm/RiZWsv/hkWFglBUyTIgyAIIfgIRjHv3r39MePM7N3LcbxAFvZ2b2bn22/mm3XMjF+HL3YW7q28YSIw8mBKoBihhhgCsoORot9d3/ywg3YowMXwNde/PzGnk2vn6PitrT+/PGeNaecg4+qNY3D43vy16A5wDDd4Aqg/ngmrjl/GoN0U5V1QquHQG3q+TPDVhVwyBffcmQGJmSVfyZk7R3SngI4JKfwDJ2+05zIg8gbiereTZRHhJ5KCMOwDFLjhoBTn2g0ghagfKeIYJDPFyibJVBtTREwq60SpYvh5++PpwatHsxSm9QRLSQpEVSd7/TYJUb49TX7gztpjjEffnoVw66+Ytovs14Yp7HaKmUXeX9rKUoMoLNW3srqI5fWn8JejrVkK0QcrkFLOgS39yoKUQe292WJ1guUHG8K2o8K00oO1BTvXoW4yasclUTgZYJY9aFNfAThX5CZRmczAV52oAPoupHhWRIUUAOoyUIlYVaAa/VbLbyiZUiyFbjQFNwiZQSGl4IDy9sO5Wrty0QLKhdZPxmgGcDo8ejn+c/6eiK9poz15Kw7Dr/vN/z6W7q++091/AQYA5mZ8GYJ9K0AAAAAASUVORK5CYII= ") left top no-repeat;
        }

        a.up {
            background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAmlJREFUeNpsU0toU0EUPfPysx/tTxuDH9SCWhUDooIbd7oRUUTMouqi2iIoCO6lceHWhegy4EJFinWjrlQUpVm0IIoFpVDEIthm0dpikpf3ZuZ6Z94nrXhhMjM3c8895977BBHB2PznK8WPtDgyWH5q77cPH8PpdXuhpQT4ifR9u5sfJb1bmw6VivahATDrxcRZ2njfoaMv+2j7mLDn93MPiNRMvGbL18L9IpF8h9/TN+EYkMffSiOXJ5+hkD+PdqcLpICWHOHc2CC+LEyA/K+cKQMnlQHJX8wqYG3MAJy88Wa4OLDvEqAEOpJd0LxHIMdHBziowSwVlF8D6QaicK01krw/JynwcKoEwZczewroTvZirlKJs5CqQ5CG8pb57FnJUA0LYCXMX5fibd+p8LWDDemcPZbzQyjvH+Ki1TlIciElA7ghwLKV4kRZstt2sANWRjYTAGzuP2hXZFpJ/GsxgGJ0ox1aoFWsDXyyxqCs26+ydmagFN/rRjymJ1898bzGzmQE0HCZpmk5A0RFIv8Pn0WYPsiu6t/Rsj6PauVTwffTSzGAGZhUG2F06hEc9ibS7OPMNp6ErYFlKavo7MkhmTqCxZ/jwzGA9Hx82H2BZSw1NTN9Gx8ycHkajU/7M+jInsDC7DiaEmo1bNl1AMr9ASFgqVu9MCTIzoGUimXVAnnaN0PdBBDCCYbEtMk6wkpQwIG0sn0PQIUF4GsTwLSIFKNqF6DVrQq+IWVrQDxAYQC/1SsYOI4pOxKZrfifiUSbDUisif7XlpGIPufXd/uvdvZm760M0no1FZcnrzUdjw7au3vu/BVgAFLXeuTxhTXVAAAAAElFTkSuQmCC ") left top no-repeat;
        }

        a.copy {
            background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAQAAAC1+jfqAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QA/4ePzL8AAAAJcEhZcwAAB2IAAAdiATh6mdsAAAAHdElNRQfmBAIGHg5kNFIeAAAA+ElEQVQoz4XQzypEcRjG8c/vzMhQ/jVFaRY2lqRmJQullNyBhbtwEe7DQnIJIjVZ2FESO1JSkxpzakIz51icczpHFp7V89bT93nfN8i0Z0Wc+0hAqunEdR0N85Z0PIuQeDMQGdo2RR1t27paFpBaFjnUxydZYELHuUJbNu070hekREgFpYIHsQOT0oJAakxLQCJY9WRO20itDMxYV8v9rTOP+ZQHgi+vorwk8a1mVNz8j4qKcYsF1MhdtmC1InZTqYjLuwrCrI3cM/T+l9BzVSH0sicJJWFau7JDV7Bj4CILBA0fTiurJxruXepmgdiutV/vTjUde4Ef7F1FoHZnGgQAAAAldEVYdGRhdGU6Y3JlYXRlADIwMjItMDQtMDJUMDY6MzA6MDcrMDA6MDCwpAChAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDIyLTA0LTAyVDA2OjMwOjA3KzAwOjAwwfm4HQAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAAASUVORK5CYII=") left top no-repeat;
        }

        a.delete {
            background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAAAAADFHGIkAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QA/4ePzL8AAAAJcEhZcwAADsQAAA7EAZUrDhsAAAAHdElNRQfmBAIIAwh4pbY9AAABLUlEQVQoz3WSQW7CQAxFfyY0lIobVS37FokzFalQNpwBicMglS44ChCPrQpU1x4HRBdNpIzzf2z9NxkIi7BdIrerMNhfS+liVL7Ca8lEHGJU3uFP+lbNoXeVjzL95zj+0NYm0DlPpuHAvzo+AnNz6CQj4E2J3WDSMQbJnP1JntE0WHl3MRZIKWGm+Rl3NYaf1mKphLPOzakwe3X9YautlFTCrTmVOaiLfpRLXHPeUVW1zRts9XDh8HtvTo0K/V3RL1sidMovSJUNW3qijsN1eUKdgBQ8EhySz1Jy9s1xHrqS55Hrg92y8CyUYndJJ2hKzsLTw9pAuBhTNK4fnKeH+423BPkUwy/P2drumH4lz7rqeEnXm46c493H8m0VgNxS6CKl6v75/6fkrxin5RfPLaqI4FGhLQAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyMi0wNC0wMlQwODowMzowMiswMDowMGqv/rQAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjItMDQtMDJUMDg6MDM6MDIrMDA6MDAb8kYIAAAAAElFTkSuQmCC") left top no-repeat;
        }

        html[dir=rtl] a {
            background-position-x: right;
        }

        .copystyle {
            float: right;
        }

        #listingParsingErrorBox {
            border: 1px solid black;
            background: #fae691;
            padding: 10px;
            display: none;
        }

        tr:nth-child(even) {
            /* background: rgb(177, 175, 175); */
            background: rgb(164, 205, 230);
        }
        h5{
            margin: 0;
            display: inline;
        }
    </style>
</head>

<body>

    <!-- <h1>Index of current directory...</h1> -->
    <h1>{{.ReqPath}}</h1>
    {{- if .AllowUpload }}
        <div style="border-bottom: 1px solid #c0c0c0;margin-bottom: 10px;
        padding-bottom: 10px; width: fit-content;">
            <h5>上传文件:</h5>

            <form method="post" enctype="multipart/form-data" style="display: inline;"><input required name="file" type="file" /><input
                value="Upload" type="submit" /></form>

        </div>
        {{- end }}
    <table width="100%">
        <tbody>
            <td>
                Name
            </td>
            <td>
                Size
            </td>
            <td>
                Date Modified
            </td>
        </tbody>
        <tr>
            <td colspan=3>
                <a class="icon up" href="{{.ParentPath}}">..</a>
            </td>
        </tr>
        {{- range .Files }}
        <tr>
            {{ if (not .IsDir) }}
            <td class=text>
                <a class="icon file" href="{{ .URL.String }}">{{ .Name }} </a>
                <div style="display: inline" class="copystyle">
                    <a class="icon copy" href="#" onclick="textCopy('{{ .ReqURL }}', '{{$.ReqPath}}' )"></a>
                </div>
            </td>
            <!-- '{{ .URL.String }}' -->
            <td class=number>{{.Size.String }}</td>
            <td class=number>{{.ModTime}}
                {{- if $.AllowDelete }}
                <div style="display: inline" class="copystyle">
                    <form action="{{$.ReqPath}}" method="POST">
                        <input type="hidden" name="_method" value="DELETE">
                        <button type="submit" name="fileName" value="{{$.ReqPath}}/{{ .Name }}" style="background-color: #f17c7c; ">delete</button>
                    </form>
                </div>
                {{- end }}
            </td>
            {{ else }}
            <td colspan=3 class=text><a class="icon dir" href="{{ .URL.String }}">{{ .Name }}</td>
            {{ end }}
        </tr>
        {{- end }}

    </table>
    <script>
        function textCopy(path, ReqPath) {
            // fullPath = reqUrl + path
            console.log("path:" + path)
            console.log("ReqPath:" + ReqPath)
            // 创建输入框
            var textarea = document.createElement('textarea');
            document.body.appendChild(textarea);
            // 隐藏此输入框
            textarea.style.position = 'absolute';
            textarea.style.clip = 'rect(0 0 0 0)';
            // 赋值
            textarea.value = path;
            // 选中
            textarea.select();
            // 复制
            document.execCommand('copy', true);
        }
        function deleteFile(path) {

        }
    </script>
</body>

</html>
`

type fileSizeBytes int64

func (f fileSizeBytes) String() string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	divBy := func(x int64) int {
		return int(math.Round(float64(f) / float64(x)))
	}
	switch {
	case f < KB:
		return fmt.Sprintf("%d", f)
	case f < MB:
		return fmt.Sprintf("%dK", divBy(KB))
	case f < GB:
		return fmt.Sprintf("%dM", divBy(MB))
	case f >= GB:
		fallthrough
	default:
		return fmt.Sprintf("%dG", divBy(GB))
	}
}

type directoryListingFileData struct {
	Name    string
	Size    fileSizeBytes
	ModTime string
	IsDir   bool
	URL     *url.URL
	ReqURL  string
}

type directoryListingData struct {
	Title       string
	ZipURL      *url.URL
	TarGzURL    *url.URL
	Files       []directoryListingFileData
	AllowUpload bool
	AllowDelete bool
	ParentPath  string
	ReqPath     string
}

type fileHandler struct {
	route       string
	path        string
	allowUpload bool
	allowDelete bool
}

var (
	directoryListingTemplate = template.Must(template.New("").Parse(directoryListingTemplateText))
)

func (f *fileHandler) serveStatus(w http.ResponseWriter, r *http.Request, status int) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(http.StatusText(status)))
	if err != nil {
		return err
	}
	return nil
}

func (f *fileHandler) serveTarGz(w http.ResponseWriter, r *http.Request, path string) error {
	w.Header().Set("Content-Type", tarGzContentType)
	name := filepath.Base(path) + ".tar.gz"
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, name))
	return tarGz(w, path)
}

func (f *fileHandler) serveZip(w http.ResponseWriter, r *http.Request, osPath string) error {
	w.Header().Set("Content-Type", zipContentType)
	name := filepath.Base(osPath) + ".zip"
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, name))
	return zip(w, osPath)
}

func getHost(r *http.Request) string {
	return r.Host + r.URL.String()
}

func (f *fileHandler) serveDir(w http.ResponseWriter, r *http.Request, osPath string) error {
	d, err := os.Open(osPath)
	if err != nil {
		return err
	}
	files, err := d.Readdir(-1)
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		// return files[i].Size() < files[j].Size()
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		} else if !files[i].IsDir() && files[j].IsDir() {
			return false
		}
		return files[i].Name() < files[j].Name()
	})

	// directoryListingTemplate, err := template.ParseFiles("list.html")
	// if err != nil {
	// 	return err
	// }

	reqHost := getHost(r)
	// fmt.Printf("url:%s\n", reqHost)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return directoryListingTemplate.Execute(w, directoryListingData{
		AllowUpload: f.allowUpload,
		AllowDelete: f.allowDelete,
		Title: func() string {
			return "Http File Server"
		}(),
		ReqPath: func() string {
			path := filepath.Clean(r.URL.String())
			return filepath.ToSlash(path)
		}(),
		ParentPath: func() string {
			path := *&r.URL.Path
			dir, _ := filepath.Split(filepath.Clean(path))
			dir = filepath.ToSlash(dir)
			if dir == "/" {
				dir = f.route
			}
			return dir
		}(),
		TarGzURL: func() *url.URL {
			url := *r.URL
			q := url.Query()
			q.Set(tarGzKey, tarGzValue)
			url.RawQuery = q.Encode()
			return &url
		}(),
		ZipURL: func() *url.URL {
			url := *r.URL
			q := url.Query()
			q.Set(zipKey, zipValue)
			url.RawQuery = q.Encode()
			return &url
		}(),
		Files: func() (out []directoryListingFileData) {
			for _, d := range files {
				name := d.Name()
				modTime := d.ModTime()
				reqURL := "http://" + reqHost + name
				fileData := directoryListingFileData{
					Name:    name,
					IsDir:   d.IsDir(),
					Size:    fileSizeBytes(d.Size()),
					ModTime: modTime.Format("2006-01-02 15:04:05"),
					ReqURL:  reqURL,
					URL: func() *url.URL {
						url := *r.URL
						url.Path = path.Join(url.Path, name)
						if d.IsDir() {
							url.Path += "/"
						}
						return &url
					}(),
				}
				out = append(out, fileData)
			}
			return out
		}(),
	})
}

func (f *fileHandler) serveUploadTo(w http.ResponseWriter, r *http.Request, osPath string) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	in, h, err := r.FormFile("file")
	if err == http.ErrMissingFile {
		w.Header().Set("Location", r.URL.String())
		w.WriteHeader(303)
	}
	if err != nil {
		return err
	}
	outPath := filepath.Join(osPath, filepath.Base(h.Filename))
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(303)
	return nil
}

// ServeHTTP is http.Handler.ServeHTTP
func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s %s %s", f.path, r.RemoteAddr, r.Method, r.URL.String())
	urlPath := r.URL.Path
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}

	rMethod := ""
	fileName := ""
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(200)
		if len(r.Form) != 0 {
			rMethod = r.Form["_method"][0]
			fileName = r.Form["fileName"][0]
		}
		if fileName != "" {
			urlPath = fileName
		}
	}
	urlPath = strings.TrimPrefix(urlPath, f.route)
	urlPath = strings.TrimPrefix(urlPath, "/"+f.route)

	osPath := strings.ReplaceAll(urlPath, "/", osPathSeparator)
	osPath = filepath.Clean(osPath)
	osPath = filepath.Join(f.path, osPath)
	info, err := os.Stat(osPath)
	// fmt.Printf("rMethod:%s, fileName:%s, osPath:%s\n", rMethod, fileName, osPath)
	switch {
	case os.IsNotExist(err):
		_ = f.serveStatus(w, r, http.StatusNotFound)
	case os.IsPermission(err):
		_ = f.serveStatus(w, r, http.StatusForbidden)
	case err != nil:
		_ = f.serveStatus(w, r, http.StatusInternalServerError)
	case !f.allowDelete && r.Method == http.MethodDelete:
		_ = f.serveStatus(w, r, http.StatusForbidden)
	case !f.allowUpload && r.Method == http.MethodPost:
		_ = f.serveStatus(w, r, http.StatusForbidden)
	case r.URL.Query().Get(zipKey) != "":
		err := f.serveZip(w, r, osPath)
		if err != nil {
			_ = f.serveStatus(w, r, http.StatusInternalServerError)
		}
	case r.URL.Query().Get(tarGzKey) != "":
		err := f.serveTarGz(w, r, osPath)
		if err != nil {
			_ = f.serveStatus(w, r, http.StatusInternalServerError)
		}
	case f.allowUpload && info.IsDir() && r.Method == http.MethodPost:
		err := f.serveUploadTo(w, r, osPath)
		if err != nil {
			_ = f.serveStatus(w, r, http.StatusInternalServerError)
		}
	case f.allowDelete && !info.IsDir() && (r.Method == http.MethodDelete || rMethod == http.MethodDelete):
		err := os.Remove(osPath)
		if err != nil {
			_ = f.serveStatus(w, r, http.StatusInternalServerError)
		}

		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)

	case info.IsDir():
		err := f.serveDir(w, r, osPath)
		if err != nil {
			_ = f.serveStatus(w, r, http.StatusInternalServerError)
		}
	default:
		http.ServeFile(w, r, osPath)
	}
}
