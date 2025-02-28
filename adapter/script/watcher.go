package script

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

func (s *Script) initWatcher() {
	s.watcher = watcher.New()
	s.watcher.SetMaxEvents(0)
	s.watcher.FilterOps(watcher.Write)

	go func() {
		if err := s.watcher.Start(time.Second); err != nil {
			s.log.Fatalf("failed to start watching scripts folder: %v", err)
			return
		}
		defer s.watcher.Close()
	}()

	go func() {
		for {
			select {
			case e := <-s.watcher.Event:
				_ = s.reload(e.Path)
			case err := <-s.watcher.Error:
				s.log.Error().Err(err).Msg("error on watching scripts folder")
			case <-s.watcher.Closed:
				return
			}
		}
	}()
}

func (s *Script) reload(path string) error {
	if dir, err := isDir(path); err != nil {
		s.log.Error().Err(err).Str("path", path).Msg("failed to check if script is a directory")
	} else if dir {
		return s.reloadDir(path)
	}
	return s.reloadFile(path)
}

func (s *Script) reloadDir(path string) error {
	s.log.Info().Str("path", path).Msg("script folder changed")

	folderScripts, err := getFolderFiles(path)
	if err != nil {
		s.log.Error().Err(err).Str("path", path).Msg("failed to get folder scripts")
		return err
	}

	loadedScripts, err := s.getScriptsByFolder(path)
	if err != nil {
		s.log.Error().Err(err).Str("path", path).Msg("failed to get loaded scripts")
		return err
	}

	diff := func(src, dst []string) string {
		for i := range src {
			exists := false
			for j := range dst {
				if src[i] == dst[j] {
					exists = true
					break
				}
			}
			if !exists {
				return src[i]
			}
		}
		return ""
	}

	if len(folderScripts) > len(loadedScripts) {
		added := diff(folderScripts, loadedScripts)
		if len(added) != 0 {
			if err := s.reloadFile(added); err != nil {
				return err
			}
		}
	} else if len(folderScripts) < len(loadedScripts) {
		removed := diff(loadedScripts, folderScripts)
		sc, ok := s.scripts.Load(removed)
		if !ok {
			s.log.Error().Str("path", removed).Msg("failed to get loaded script")
			return errors.New("failed to get loaded script")
		}

		s.log.Info().Str("path", removed).Msg("script removed")

		s.scripts.Delete(removed)
		sc.(*script).close()
	}

	return nil
}

func (s *Script) reloadFile(path string) error {
	sc, exists := s.scripts.Load(path)
	if exists {
		s.scripts.Delete(path)
		sc.(*script).close()
	}

	if err := s.loadScript(path); err != nil {
		s.log.Error().Err(err).Str("path", path).Msg("failed to load script")
		return err
	}

	if exists {
		s.log.Info().Str("path", path).Msg("script reloaded")
	} else {
		s.log.Info().Str("path", path).Msg("script loaded")
	}

	return nil
}

func (s *Script) getScriptsByFolder(folder string) (scripts []string, err error) {
	var matched bool
	scripts = make([]string, 0, 1)

	s.scripts.Range(func(k, v interface{}) bool {
		matched, err = filepath.Match(filepath.Join(folder, "/*"), k.(string))
		if err != nil {
			return false
		}
		if matched {
			scripts = append(scripts, v.(*script).path)
		}
		return true
	})

	return
}

func isDir(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

func getFolderFiles(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	scripts := make([]string, 0, len(files))
	for _, d := range files {
		if d.IsDir() {
			continue
		}

		f, err := filepath.Abs(filepath.Join(path, d.Name()))
		if err != nil {
			return nil, err
		}

		scripts = append(scripts, f)
	}

	return scripts, nil
}
