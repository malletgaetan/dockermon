package persist

// import (
//     "encoding/json"
//     "os"
//     "sync"
//     "time"
// )

// type Storage struct {
//     lastHandledEvent time.Time         `json:"last_handled_event"`
// }

// type Manager struct {
//     filePath string
// 	data     Storage
//     mu       sync.RWMutex
// }

// func NewManager(path string) (*Manager, error) {
//     m := &Manager{
//         filePath: path,
//     }

//     if err := m.load(); err != nil {
//         if !os.IsNotExist(err) {
//             return nil, err
//         }
//         if err := m.save(); err != nil {
//             return nil, err
//         }
//     }

//     return m, nil
// }

// func (m *Manager) load() error {
//     m.mu.Lock()
//     defer m.mu.Unlock()

//     data, err := os.ReadFile(m.filePath)
//     if err != nil {
//         return err
//     }

//     return json.Unmarshal(data, &m.data)
// }

// func (m *Manager) save() error {
//     m.mu.Lock()
//     defer m.mu.Unlock()

//     data, err := json.MarshalIndent(m.data, "", "  ")
//     if err != nil {
//         return err
//     }

//     return os.WriteFile(m.filePath, data, 0644)
// }

// func (m *Manager) UpdateLastHandledEvent() error {
//     m.mu.Lock()
//     m.data.lastHandledEvent = time.Now()
//     m.mu.Unlock()
//     return m.save()
// }

// func (m *Manager) GetLastHandledEvent() time.Time {
//     m.mu.RLock()
//     defer m.mu.RUnlock()
//     return m.data.lastHandledEvent
// }
