package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func ReadCSVFiles(directory string) ([]string, error) {
	files, err := func() ([]fs.FileInfo, error) {
		f, err := os.Open(directory)
		if err != nil {
			return nil, fmt.Errorf("无法打开目录 %s: %w", directory, err)
		}
		defer f.Close()

		list, err := f.Readdir(-1)
		if err != nil {
			return nil, fmt.Errorf("无法读取目录 %s 的内容: %w", directory, err)
		}

		sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
		return list, nil
	}()

	if err != nil {
		return nil, err
	}

	var csvFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			csvFiles = append(csvFiles, filepath.Join(directory, file.Name()))
		}
	}

	if len(csvFiles) == 0 {
		return nil, fmt.Errorf("在目录 %s 中未找到CSV文件", directory)
	}

	return csvFiles, nil
}

func MoveFilesToBackup(srcFolder, backupFolder string) error {
	// 确保备份文件夹存在
	if _, err := os.Stat(backupFolder); os.IsNotExist(err) {
		err := os.Mkdir(backupFolder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建备份文件夹失败: %v", err)
		}
	}

	// 读取源文件夹中的所有文件
	files, err := filepath.Glob(filepath.Join(srcFolder, "*"))
	if err != nil {
		return fmt.Errorf("读取源文件夹失败: %v", err)
	}

	for _, file := range files {
		// 获取文件的基础名称
		fileName := filepath.Base(file)

		// 构造目标路径
		destPath := filepath.Join(backupFolder, fileName)

		// 移动文件
		err := moveFile(file, destPath)
		if err != nil {
			return fmt.Errorf("移动文件失败: %v", err)
		}
	}

	return nil
}

// 移动文件的辅助函数
func moveFile(srcPath, destPath string) error {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 删除源文件
	err = os.Remove(srcPath)
	if err != nil {
		return fmt.Errorf("删除源文件失败: %v", err)
	}

	return nil
}
