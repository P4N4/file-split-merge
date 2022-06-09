package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

//1 мегабайт в байтах для подсчета удаления ненужных нулей и деления файла
const a = 1048576

type fileinfo struct {
	size     string
	nameType string
}

//функция деления файла на фрагменты
func splitFile(w http.ResponseWriter, r *http.Request) {

	var splittedFile [][]byte
	//загружаем файл через форму
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	//считываем и записываем файл в папку
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile("./upload/"+handler.Filename, fileBytes, 0755)
	if err != nil {
		fmt.Println(err)
	}

	//создаем папку для хранения фрагментов файла
	err = os.Mkdir("./splitted/"+handler.Filename, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
	}

	pathOfFile := "./upload/" + handler.Filename
	pathToSave := "./splitted/" + handler.Filename + "/"

	//считывеам загруженный файл
	fileUploadded, err := ioutil.ReadFile(pathOfFile)
	if err != nil {
		fmt.Println(err)
	}

	//считаем колличество полных мегабайт файла и сколько требуется дописать нулей
	fullMBs := len(fileUploadded)/a + 1
	lastWeight := fullMBs*a - len(fileUploadded) - 1

	//добавляем нули
	for i := 0; i <= lastWeight; i++ {
		fileUploadded = append(fileUploadded, 0)
	}

	//делим файл на фрагменты 1 мб
	for i := 0; i <= fullMBs-1; i++ {
		splittedFile = append(splittedFile, fileUploadded[a*i:a*(i+1)])
	}

	fileSizeToString := strconv.Itoa(int(handler.Size))
	//записываем фрагменты
	for i := 0; i <= len(splittedFile)-1; i++ {
		err = ioutil.WriteFile(pathToSave+strconv.Itoa(i)+"."+fileSizeToString+"."+handler.Filename, splittedFile[i], 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Fprintf(w, "Файл успешно загружен на сервер\n")

}

//функция, собирающая файл из фрагментов
func mergeFiles(w http.ResponseWriter, r *http.Request) {

	var newFile []byte
	var files [][]byte
	var fileInfo fileinfo

	//получаем название файла из url
	urlName := string(r.URL.RequestURI())
	newUrl := urlName[10:]
	veryNewUrl, err := url.QueryUnescape(newUrl)
	if err != nil {
		fmt.Println(err)
	}

	//считываем название файлов в папке, где хранятся фрагменты нужного нам файла, в названиях
	//хранятся данные о ненужных нулях, порядок файлов и исходное название файла
	filesDir, err := os.ReadDir("./splitted/" + veryNewUrl)
	if err != nil {
		fmt.Println(err)
	}

	//полученные данные преобразовываем в читаемый вид
	fileName := filesDir[0].Name()
	newFileName := strings.Split(fileName, ".")
	fileInfo.nameType = strings.Join(newFileName[2:], ".")
	fileInfo.size = newFileName[1]

	//считываем фрагменты и записываем их в срез вида [][]
	for i := 0; ; i++ {

		f, err := ioutil.ReadFile("./splitted/" + fileInfo.nameType + "/" + strconv.Itoa(i) + "." + fileInfo.size + "." + fileInfo.nameType)
		if err != nil {
			fmt.Println(err)
			break
		}
		files = append(files, f)
	}

	//преобразуем считанные ранее фрагменты срез вида []
	for i := 0; i <= len(files)-1; i++ {
		newFile = append(newFile, files[i]...)
	}

	//отрезаем ненужные нам нули
	lenBytes, err := strconv.Atoi(fileInfo.size)
	if err != nil {
		fmt.Println(err)
	}

	newFile = newFile[:lenBytes]
	err = ioutil.WriteFile("./merged/"+fileInfo.nameType, newFile, 0755)
	if err != nil {
		fmt.Println(err)
	}

	//окно скачивания файла
	http.ServeFile(w, r, "./merged/"+veryNewUrl)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "index.html")

	err := os.Mkdir("./splitted/", 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
	}

	err = os.Mkdir("./merged/", 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
	}

	err = os.Mkdir("./upload/", 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/download/", mergeFiles)
	mux.HandleFunc("/upload", splitFile)
	mux.HandleFunc("/", indexHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error creating http server: ", err)
	}
}
