package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() 에러 = %v, want nil (파일이 없는 건 정상 상태여야 함)", err)
	}
	if s.Has("아무거나") {
		t.Errorf("처음 로드한 Store에 뭔가 이미 있음 — 비어있어야 함")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() 에러 = %v, want nil", err)
	}
	if s.Has("아무거나") {
		t.Errorf("빈 파일을 로드했는데 뭔가 있음 — 비어있어야 함")
	}
}

// TestLoad_InvalidJSON은 "잘못된 입력이 실제로 실패(에러)를 반환하는지"를
// 검사하는 케이스입니다. 앞의 두 테스트(파일 없음/빈 파일)는 둘 다
// "정상 상태"로 취급돼 에러가 nil이어야 하지만, 이 케이스는 반대로
// err이 반드시 nil이 아니어야 통과합니다.
func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(path, []byte("이건 JSON이 아님{{{"), 0o644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load()가 깨진 JSON에도 에러 없이 성공함 — 에러를 반환해야 함")
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seen.json")
	if err := os.WriteFile(path, []byte(`{"2026000449": true}`), 0o644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() 에러 = %v, want nil", err)
	}
	if !s.Has("2026000449") {
		t.Errorf("기존에 저장된 ID를 못 읽어옴")
	}
	if s.Has("없는ID") {
		t.Errorf("저장 안 된 ID가 있다고 나옴")
	}
}

func TestMarkSeenThenSaveThenReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seen.json")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() 에러 = %v, want nil", err)
	}

	if s.Has("new-id") {
		t.Fatalf("저장 전인데 이미 있다고 나옴")
	}
	s.MarkSeen("new-id")
	if !s.Has("new-id") {
		t.Fatalf("MarkSeen 직후인데 Has가 false")
	}

	if err := s.Save(); err != nil {
		t.Fatalf("Save() 에러 = %v, want nil", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("재로드 Load() 에러 = %v, want nil", err)
	}
	if !reloaded.Has("new-id") {
		t.Errorf("디스크에 저장된 걸 다시 불러왔는데 없음 — Save/Load가 안 맞물림")
	}
}

// TestSave_DirectoryDoesNotExist도 "실패를 제대로 반환하는지"를 검사하는
// 케이스입니다. 부모 디렉토리가 없는 경로로 Save를 호출하면 os.WriteFile이
// 실패해야 하고, Save는 그 에러를 그대로 호출자에게 전파해야 합니다.
func TestSave_DirectoryDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "존재하지않는디렉토리", "seen.json")

	s, err := Load(path) // 디렉토리가 없어도 "파일 없음"과 동일하게 취급되어 성공해야 함
	if err != nil {
		t.Fatalf("Load() 에러 = %v, want nil", err)
	}

	s.MarkSeen("id")
	if err := s.Save(); err == nil {
		t.Fatal("Save()가 존재하지 않는 디렉토리에도 에러 없이 성공함 — 에러를 반환해야 함")
	}
}
