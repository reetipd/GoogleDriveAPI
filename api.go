package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func getDriveService() (*drive.Service, error) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		fmt.Printf("Unable to read credentials.json file. Err: %v\n", err)
		return nil, err
	}
	// ConfigFromJSON uses a Google Developers Console client_credentials.json file to construct a config. client_credentials.json
	// If you want to modifyt this scope, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)

	if err != nil {
		return nil, err
	}

	client := getClient(config)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))

	if err != nil {
		fmt.Printf("Cannot create the Google Drive service: %v\n", err)
		return nil, err
	}

	return service, err
}

// Retrieve a token, saves the token, then returns the generated client.
//Uses config to retrieve a token
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)

	//token hasn't been generated
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	// fmt.Println("Token", &tok)
	err = json.NewDecoder(f).Decode(tok)
	// fmt.Println("Err",err)
	return tok, err
}

//Uses config to request token
// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	fmt.Println("Paste Authrization code here :")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func createFolder(service *drive.Service, name string, parentId string) (*drive.File, error) {
	f := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	folder, err := service.Files.Create(f).Fields("webViewLink, id").Do()

	if err != nil {
		fmt.Println("Could not create folder: " + err.Error())
		return nil, err
	}

	return folder, nil
}

func renameFolder(service *drive.Service, name string, Id string) (*drive.File, error){
	f := &drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:                         name,
	}

	file, err := service.Files.Update(Id, f).Do()

	if err != nil{
		log.Printf("Could not rename folder: "+err.Error())
		return nil, err
	}
	return file, nil
}

func callCreateFolder(srv *drive.Service){
	// Create Folder
	folder, err := createFolder(srv, "New Folder", "root")

	if err != nil {
		panic(fmt.Sprintf("Could not create folder: %v\n", err))
	}

	// link := folder.Header.Get("webViewLink")

	fmt.Println("Created folder :->", folder.Name, folder.Id )
}

func callRenameFolder(srv *drive.Service){
	//taking input from user for renaming folder
	fmt.Println("Enter new name for folder: ")
	var newName string
	fmt.Scanln(&newName)

	//Rename folder
	folder, err := renameFolder(srv,newName,"1BPHSIcl2gu3d3zp8bBl4R5L8YmhJxoRR")
	if err != nil {
		panic(fmt.Sprintf("Could not rename folder: %v\n", err))
	}

	fmt.Println("Renamed folder name and ID:->", folder.Name, folder.Id)
}

func getFolderLink(service *drive.Service) (string){
	// fileLink, err := service.Files.List().Fields("webViewLink").Do()
	fileLink, err := service.Files.Get("12MJM0cOQUXbrP3pyaX-hVe22U6yJY27T").Fields("webViewLink").Do()
	if err != nil{
		fmt.Println("Could not get folder link: "+ err.Error())
	}
	fmt.Println("FileLink", fileLink.WebViewLink)
	return fileLink.WebViewLink
}

func callCreateFiles(srv *drive.Service, name string, mimeType string, parentId string) (*drive.File, error){
	fmt.Printf(mimeType)
	file := &drive.File{
		MimeType: mimeType,
		Name: name,
		Parents: []string{parentId},
	}

	file, err := srv.Files.Create(file).Do()
	
	if err != nil {
		fmt.Println("Could not create file: " + err.Error())
		return nil, err
	}

	return file, nil
}

func InsertPermission(srv *drive.Service, fileId string, permissionType string, email string, permissionRole string) (*drive.Permission, error){
	p := &drive.Permission{
		Type: permissionType,
		Role: permissionRole,
		EmailAddress: email,
	}

	filePermission, err := srv.Permissions.Create(fileId,p).TransferOwnership(false).Do()
	if err != nil{
		fmt.Printf("Error while creating permission for file: %v\n", err)
		return nil,err
	}
	fmt.Println("Persmissions....", filePermission.Role)

	return filePermission, nil

}

func GetList(srv *drive.Service, fileId string) (error){
	r, err := srv.Permissions.List(fileId).Fields("*").Do()
	if err != nil{
		fmt.Printf("Error while getting list: %v\n", err)
		return err
	}
	permissionListID := r.Permissions
	fmt.Println("data..", permissionListID, r.Kind)

	fmt.Printf("User having access to file %v is:\n",fileId)
	for _, val := range(permissionListID){
		fmt.Println("Email Address: ",val.EmailAddress)
	}
	return nil
}




func GetAllFiles(srv *drive.Service, folderID string) ([]*drive.File,error){
	var fs []*drive.File
	// pageToken := ""

	//query for finding the files in defined folder and which has not been added in bin
	query := "parents in '"+ folderID + "' and trashed=false" 
	q := srv.Files.List().Q(query)
	//if we have a pagetoken set, apply it to the query
	// if pageToken != ""{
	// 	q = q.PageToken(pageToken)
	// }

	r,err := q.Do()

	if err != nil{
		fmt.Printf("Error occurred: %v\n",err.Error())
		return fs,err
	}

	fs = append(fs, r.Files...)
	// pageToken = r.NextPageToken
	// if pageToken == "" {
	// 	break
	// }
	// fmt.Println("q..",q)
	// fmt.Println("fs..",len(fs))

	for i := 0; i<len(fs); i++ {
		fmt.Printf("Filename: %v, FileType: %v, FileID:%v\n",fs[i].Name, fs[i].MimeType, fs[i].Id)
	}
	return fs, nil
}



func main() {

	// Get the Google Drive service
	srv, _ := getDriveService()

	//1.Create empty folder
	// callCreateFolder(srv)
	
	//2.Rename folder
	// callRenameFolder(srv)


	//3. Get folderLink and obtain FolderId based on link
	folderLink := getFolderLink(srv)
	splits := strings.Split(folderLink,"/")
	folderID := splits[len(splits)-1] 


	// op,_:=srv.About.Get().Fields("user").Do()
	// fmt.Println("op..",op)



	// //4.Create docs and sheets
	// // f, err := os.Open("sample.doc")
	// // if err != nil {
	// // 	fmt.Printf("cannot open file: %v", err)
	// // }

	// // defer f.Close()
	// // gsheet,err := os.Open("sample.xlsx")
	// // if err != nil{
	// // 	fmt.Printf("cannot open file: %v", err)
	// // }
	// // defer gsheet.Close()

	// //4.1...DOC
	// docsFile, err := callCreateFiles(srv,"Untitled", "application/vnd.google-apps.document",folderID)

	// if err != nil {
	// 	fmt.Printf("Could not create docs file: %v\n", err)
	// }

	// fmt.Printf("Docs File '%s' created", docsFile.Name)
	// fmt.Printf("\nDocs File Id: '%s' ", docsFile.Id)

	// //4.2...GSHEET
	// gsheetFile, err := callCreateFiles(srv, "Untitled","application/vnd.google-apps.spreadsheet",folderID)
	// if err != nil {
	// 	fmt.Printf("Could not create google sheet file: %v\n", err)
	// }

	// fmt.Printf("GSheet File '%s' created", gsheetFile.Name)
	// fmt.Printf("\n GSheet File Id: '%s' ", gsheetFile.Id)


	//5.Provide permission to user
	// InsertPermission(srv, folderID, "user","ritipradhan1234@gmail.com","writer")

	//6.get list of users having permission
	// GetList(srv, folderID)

	//7.get folder and its file and file type
	GetAllFiles(srv, folderID)

}
