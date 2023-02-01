package files

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// 获取文件的MD5值
func CalcFileMD5(filePath string) (string, error) {
	// 打开文件
	value := ""
	f, err := os.Open(filePath)
	if err != nil {
		return value, fmt.Errorf("计算文件MD5失败, 无法打开文件! 文件路径: %s 错误信息: %v", filePath, err)
	}
	defer f.Close()

	md5Handle := md5.New()
	_, err = io.Copy(md5Handle, f)
	value = hex.EncodeToString(md5Handle.Sum(nil))
	return value, err
}

// 删除文件或目录
func Delete(path string) error {
	err := os.RemoveAll(path)
	return err
}

// 复制目录

// 判断目录是否存在,目录存在返回true
func PathIsExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// 判断目录是否存在
func ExistDir(dirname string) bool {
	fi, err := os.Stat(dirname)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}

// 检查目录或文件是否存在
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// 目录不存在则创建目录
func CreateDirIfNotExist(dirPath string, permMode os.FileMode) error {
	if Exists(dirPath) {
		return nil
	}
	err := os.MkdirAll(dirPath, permMode)
	return err
}

// 获取给定目录下的所有文件
func GetFiles(dirPath string) ([]string, error) {
	var fileArray []string
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			fileArray = append(fileArray, path)
		}
		return nil
	})
	return fileArray, err
}

// 根据glob规则获取指定目录下匹配的所有文件
func SearchFileInPath(dirPath string, fileNameGlobReg string) ([]string, error) {
	var goalFileArrary []string
	fileArray, err := GetFiles(dirPath)
	if err == nil {
		for _, eachFilePath := range fileArray {
			fileBaseName := filepath.Base(eachFilePath)
			match, err := filepath.Match(fileNameGlobReg, fileBaseName)
			if err != nil {
				return goalFileArrary, err
			} else if match {
				goalFileArrary = append(goalFileArrary, eachFilePath)
			}
		}
	}
	return goalFileArrary, err
}

// 拷贝文件夹
func CopyDirectory(scrDir, dstDir string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dstDir, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateDirIfNotExist(destPath, fs.ModePerm); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Type()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Type()); err != nil {
				return err
			}
		}
	}
	return nil
}

// 拷贝文件
func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

// 逐行读取文件
func ReadFileAsLines(filePath string) ([]string, error) {
	var resultSlice []string
	f, err := os.Open(filePath)
	if err != nil {
		return resultSlice, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 获取最后一行内容后退出循环
				resultSlice = append(resultSlice, line)
				break
			}
			return resultSlice, fmt.Errorf("读取文件内容失败,换行符读取异常! ERROR: %s", err.Error())
		}
		resultSlice = append(resultSlice, line)
	}
	return resultSlice, err
}

// 逐行读取文件为字节切片
func ReadFileAsByteSlice(filePath string) ([]byte, error) {
	bytes, err := os.ReadFile(filePath)
	return bytes, err
}

// 将字节切片数据逐行写入文件
func WriteByteSlice2File(filePath string, byteSlice []byte) error {
	err := os.WriteFile(filePath, byteSlice, 0666)
	return err
}

// 解压gzip文件
func DeCompressGzip(gzipFile, dest string) error {
	// 打开准备解压的gzip文件
	srcFile, err := os.Open(gzipFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 将打开的文件先解压
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()

	// 通过 gr 创建 tar.Reade
	tr := tar.NewReader(gr)
	// 现在已经获得了 tar.Reader 结构了,只需要循环里面的数据写入文件就可以了
	for {
		hdr, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		// 处理下保存路径,将要保存的目录加上 header 中的 Name
		// 这个变量保存的有可能是目录,有可能是文件,所以就叫 FileDir了
		dstFileDir := filepath.Join(dest, hdr.Name)

		// 根据 header 的 Typeflag 字段,判断文件的类型
		switch hdr.Typeflag {
		case tar.TypeDir: // 如果是目录时候,创建目录
			// 判断下目录是否存在,不存在就创建
			if b := ExistDir(dstFileDir); !b {
				// 使用 MkdirAll 不使用 Mkdir ,就类似 Linux 终端下的 mkdir -p,
				// 可以递归创建每一级目录
				if err := os.MkdirAll(dstFileDir, 0775); err != nil {
					return err
				}
			}
		case tar.TypeReg: // 如果是文件就写入到磁盘
			// 创建一个可以读写的文件,权限就使用 header 中记录的权限
			// 因为操作系统的 FileMode 是 int32 类型的,hdr 中的是 int64,所以转换下
			file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tr)
			if err != nil {
				return err
			}
			// 不要忘记关闭打开的文件,因为它是在 for 循环中,不能使用 defer
			// 如果想使用 defer 就放在一个单独的函数中
			file.Close()
		}

	}
}

// 将文件打包为tar
func CompressTar(source, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

// 解开tar包
func UnTar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

// this is the default sort order of golang ReadDir
func SortFileNameAscend(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
}

func SortFileNameDescend(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})
}
