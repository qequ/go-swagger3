package gomod

import (
	"github.com/mikunalpha/go-module"
	"github.com/parvez3019/go-swagger3/parser/model"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Parser interface {
	Parse() error
}

type parser struct {
	model.Utils
}

func NewParser(utils model.Utils) Parser {
	return &parser{
		Utils: utils,
	}
}

// Parse parse go.mod info
func (p *parser) Parse() error {
	log.Info("Parsing GoMod Info ...")
	b, err := ioutil.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	goMod, err := module.Parse(b)
	if err != nil {
		return err
	}
	for i := range goMod.Requires {
		if err = p.parseGoModFilePackages(goMod.Requires[i].Path, goMod.Requires[i].Version); err != nil {
			return err
		}
	}
	if p.RunInDebugMode {
		for i := range p.KnownPkgs {
			p.Debugf(p.KnownPkgs[i].Name, "->", p.KnownPkgs[i].Path)
		}
	}
	return nil
}

func (p *parser) parseGoModFilePackages(pkgName string, version string) error {
	pathRunes := []rune{}
	for _, v := range pkgName {
		if !unicode.IsUpper(v) {
			pathRunes = append(pathRunes, v)
			continue
		}
		pathRunes = append(pathRunes, '!')
		pathRunes = append(pathRunes, unicode.ToLower(v))
	}
	pkgPath := filepath.Join(p.GoModCachePath, string(pathRunes)+"@"+version)
	pkgName = filepath.ToSlash(pkgName)
	p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
		Name: pkgName,
		Path: pkgPath,
	})
	p.KnownNamePkg[pkgName] = &p.KnownPkgs[len(p.KnownPkgs)-1]
	p.KnownPathPkg[pkgPath] = &p.KnownPkgs[len(p.KnownPkgs)-1]

	return filepath.Walk(pkgPath, p.walkerFunc(pkgName, pkgPath))
}

func (p *parser) walkerFunc(pkgName string, pkgPath string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
				return nil
			}
			fns, err := filepath.Glob(filepath.Join(path, "*.go"))
			if len(fns) == 0 || err != nil {
				return nil
			}
			// p.debug(path)
			name := filepath.Join(pkgName, strings.TrimPrefix(path, pkgPath))
			name = filepath.ToSlash(name)
			p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
				Name: name,
				Path: path,
			})
			p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		}
		return nil
	}
}
