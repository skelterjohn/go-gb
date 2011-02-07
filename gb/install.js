// This is a Windows Script Host script to build gb-go on windows without gnu
//
// To run this file on windows just double click it or navigate to the current
// directory in command prompt and run 'cscript install.js'

var fso = WScript.CreateObject("Scripting.FileSystemObject");
var shell = WScript.CreateObject("WScript.Shell")

function getSourceFiles() {	
	var cd = fso.GetFolder(".");
	var cdFiles = cd.Files;

	var e = new Enumerator(cdFiles);

	var srcs = "";

	for (; !e.atEnd(); e.moveNext()){
		var fileName = e.item().Name;
		var ext = fileName.substr(fileName.length-3)

		// TODO: Find which package the file is in
		//     Then ignore it if its not in 'main'
		// For now we just ignore doc.go
		if(fileName == "doc.go") {
			continue
		}

		if(ext == ".go") {
			srcs += fileName + " ";
		}
	}
	
	return srcs;
}

function runAndWait(cmd) {
	var running = shell.Exec(cmd)
	while(running.status == 0)
		WScript.Sleep(10)
	result = running.StdOut.ReadAll()
	if(result) {
		WScript.Echo(result)
	}
}

var srcs = getSourceFiles()
var gobin = shell.ExpandEnvironmentStrings("%GOBIN%")
var goarch = shell.ExpandEnvironmentStrings("%GOARCH%")
var target = "gb.exe"

if(goarch == "386") {
	runAndWait(gobin + "/8g -o _compiled_ " + srcs)
	runAndWait(gobin + "/8l -o " + target + " _compiled_")
} else if(goarch == "amd64") {
	runAndWait(gobin + "/6g -o _compiled_ " + srcs)
	runAndWait(gobin + "/6l -o " + target)
}

var gb = fso.GetFile(target)
gb.Copy(gobin + "/" + target)

