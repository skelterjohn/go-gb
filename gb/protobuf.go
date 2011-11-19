package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func GoForProto(protosrc string) (gosrc string) {
	base := protosrc[:len(protosrc)-len(".proto")]
	gosrc = base + ".pb.go"
	return
}

func GenerateProtobufSource(this *Package) (err os.Error) {
	for _, pbs := range this.ProtoSrcs {
		args := []string{"protoc", "--go_out=.", pbs}
		if Verbose {
			fmt.Println(args)
		}
		err = RunExternal(ProtocCMD, this.Dir, args)
		if err != nil {
			return
		}

		gosrc := GoForProto(pbs)

		var protopkg string
		protopkg, _, _, _, _, _, err = GetDeps(filepath.Join(this.Dir, gosrc))
		if err != nil {
			return
		}

		this.PkgSrc[protopkg] = append(this.PkgSrc[protopkg], gosrc)
	}
	return
}
