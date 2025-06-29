package shmem

/*
#include <windows.h>
#include "../hwisenssm2.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/shayne/hwinfo-streamdeck/internal/hwinfo/mutex"
	"github.com/shayne/hwinfo-streamdeck/internal/hwinfo/util"
	"golang.org/x/sys/windows"
)

var buf = make([]byte, 200000)

func copyBytes(addr uintptr) []byte {
	headerLen := C.sizeof_HWiNFO_SENSORS_SHARED_MEM2

	var d []byte
	dh := (*reflect.SliceHeader)(unsafe.Pointer(&d))

	dh.Data = addr
	dh.Len, dh.Cap = headerLen, headerLen

	cheader := C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&d[0]))
	fullLen := int(cheader.dwOffsetOfReadingSection + (cheader.dwSizeOfReadingElement * cheader.dwNumReadingElements))

	if fullLen > cap(buf) {
		buf = append(buf, make([]byte, fullLen-cap(buf))...)
	}

	dh.Len, dh.Cap = fullLen, fullLen

	copy(buf, d)

	return buf[:fullLen]
}

// ReadBytes copies bytes from global shared memory with retries
func ReadBytes() ([]byte, error) {
	maxRetries := 5
	retryDelay := 1 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := mutex.Lock()
		if err != nil {
			lastErr = fmt.Errorf("failed to acquire mutex: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		hnd, err := openFileMapping()
		if err != nil {
			mutex.Unlock()
			lastErr = fmt.Errorf("failed to open file mapping: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		addr, err := mapViewOfFile(hnd)
		if err != nil {
			windows.CloseHandle(windows.Handle(unsafe.Pointer(hnd)))
			mutex.Unlock()
			lastErr = fmt.Errorf("failed to map view of file: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		// Check if the shared memory is initialized
		var d []byte
		dh := (*reflect.SliceHeader)(unsafe.Pointer(&d))
		dh.Data = addr
		dh.Len, dh.Cap = C.sizeof_HWiNFO_SENSORS_SHARED_MEM2, C.sizeof_HWiNFO_SENSORS_SHARED_MEM2
		cheader := C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&d[0]))

		if cheader.dwSignature != 0x53695748 { // "HWiS"
			unmapViewOfFile(addr)
			windows.CloseHandle(windows.Handle(unsafe.Pointer(hnd)))
			mutex.Unlock()
			lastErr = fmt.Errorf("invalid shared memory signature: 0x%x", cheader.dwSignature)
			time.Sleep(retryDelay)
			continue
		}

		data := copyBytes(addr)
		unmapViewOfFile(addr)
		windows.CloseHandle(windows.Handle(unsafe.Pointer(hnd)))
		mutex.Unlock()
		return data, nil
	}

	return nil, fmt.Errorf("failed to read shared memory after %d retries: %v", maxRetries, lastErr)
}

func openFileMapping() (C.HANDLE, error) {
	lpName := C.CString(C.HWiNFO_SENSORS_MAP_FILE_NAME2)
	defer C.free(unsafe.Pointer(lpName))

	hnd := C.OpenFileMapping(syscall.FILE_MAP_READ, 0, lpName)
	if hnd == C.HANDLE(C.NULL) {
		errstr := util.HandleLastError(uint64(C.GetLastError()))
		return nil, fmt.Errorf("OpenFileMapping: %w", errstr)
	}

	return hnd, nil
}

func mapViewOfFile(hnd C.HANDLE) (uintptr, error) {
	addr, err := windows.MapViewOfFile(windows.Handle(unsafe.Pointer(hnd)), C.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("MapViewOfFile: %w", err)
	}

	return addr, nil
}

func unmapViewOfFile(ptr uintptr) error {
	err := windows.UnmapViewOfFile(ptr)
	if err != nil {
		return fmt.Errorf("UnmapViewOfFile: %w", err)
	}
	return nil
}
