package main

import (
	"path/filepath"
)

func GoForYacc(yaccsrc string) (gosrc string) {
	gosrc = yaccsrc + ".go"
	return
}

func GenerateGoyaccSource(this *Package) (err error) {
	for _, ys := range this.YaccSrcs {
		base := ys[:len(ys)-len(".y")]
		gosrc := GoForYacc(ys)
		args := []string{"goyacc", "-o", gosrc, "-p", base, ys}

		err = RunExternal(GoYaccCMD, this.Dir, args)
		if err != nil {
			return
		}

		var pkg string
		pkg, _, _, _, _, _, err = GetDeps(filepath.Join(this.Dir, gosrc))
		if err != nil {
			return
		}

		this.PkgSrc[pkg] = append(this.PkgSrc[pkg], gosrc)

		// this probably can't actually happen
		if this.Name == "" {
			this.Name = pkg
		}
	}
	return
}
