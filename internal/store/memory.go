package store

import (
	"sync"

	"go05-charity-project/internal/model"
)

type Snapshot struct {
	Users         []model.User         `json:"users"`
	Orgs          []model.Organization `json:"orgs"`
	Projects      []model.Project      `json:"projects"`
	Donations     []model.Donation     `json:"donations"`
	Ledgers       []model.LedgerEntry  `json:"ledgers"`
	Expenses      []model.Expense      `json:"expenses"`
	Progress      []model.ProgressReport `json:"progress"`
	Follows       []model.Follow       `json:"follows"`
	Comments      []model.Comment      `json:"comments"`
	Receipts      []model.Receipt      `json:"receipts"`
	Notifications []model.Notification `json:"notifications"`
	Audits        []model.AuditLog     `json:"audits"`
}

type MemoryStore struct {
	mu      sync.RWMutex
	hook    func()
	idGen   func(prefix string) string
	codeGen func(n int) string

	users         map[string]model.User
	usersByName   map[string]string
	orgs          map[string]model.Organization
	projects      map[string]model.Project
	donations     map[string]model.Donation
	ledgers       map[string]model.LedgerEntry
	expenses      map[string]model.Expense
	progress      map[string]model.ProgressReport
	follows       map[string]model.Follow
	comments      map[string]model.Comment
	receipts      map[string]model.Receipt
	receiptsByCode map[string]string
	notifications map[string]model.Notification
	audits        map[string]model.AuditLog
}

func NewMemoryStore(idGen func(string) string, persist func()) *MemoryStore {
	if idGen == nil {
		idGen = defaultIDGenerator
	}
	m := &MemoryStore{
		hook:           persist,
		idGen:          idGen,
		codeGen:        randomCode,
		users:          map[string]model.User{},
		usersByName:    map[string]string{},
		orgs:           map[string]model.Organization{},
		projects:       map[string]model.Project{},
		donations:      map[string]model.Donation{},
		ledgers:        map[string]model.LedgerEntry{},
		expenses:       map[string]model.Expense{},
		progress:       map[string]model.ProgressReport{},
		follows:        map[string]model.Follow{},
		comments:       map[string]model.Comment{},
		receipts:       map[string]model.Receipt{},
		receiptsByCode: map[string]string{},
		notifications:  map[string]model.Notification{},
		audits:         map[string]model.AuditLog{},
	}
	return m
}

func (m *MemoryStore) SetPersistHook(fn func()) { m.hook = fn }

func (m *MemoryStore) persist() {
	if m.hook != nil {
		m.hook()
	}
}

func (m *MemoryStore) NewID(prefix string) string { return m.idGen(prefix) }

func (m *MemoryStore) NewCode(n int) string { return m.codeGen(n) }

func (m *MemoryStore) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshotNoLock()
}

func (m *MemoryStore) snapshotNoLock() Snapshot {
	s := Snapshot{
		Users:         values(m.users),
		Orgs:          values(m.orgs),
		Projects:      values(m.projects),
		Donations:     values(m.donations),
		Ledgers:       values(m.ledgers),
		Expenses:      values(m.expenses),
		Progress:      values(m.progress),
		Follows:       values(m.follows),
		Comments:      values(m.comments),
		Receipts:      values(m.receipts),
		Notifications: values(m.notifications),
		Audits:        values(m.audits),
	}
	return s
}

func (m *MemoryStore) ReplaceAll(s Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = map[string]model.User{}
	m.usersByName = map[string]string{}
	m.orgs = map[string]model.Organization{}
	m.projects = map[string]model.Project{}
	m.donations = map[string]model.Donation{}
	m.ledgers = map[string]model.LedgerEntry{}
	m.expenses = map[string]model.Expense{}
	m.progress = map[string]model.ProgressReport{}
	m.follows = map[string]model.Follow{}
	m.comments = map[string]model.Comment{}
	m.receipts = map[string]model.Receipt{}
	m.receiptsByCode = map[string]string{}
	m.notifications = map[string]model.Notification{}
	m.audits = map[string]model.AuditLog{}
	for _, v := range s.Users {
		m.users[v.ID] = v
		m.usersByName[v.Username] = v.ID
	}
	for _, v := range s.Orgs {
		m.orgs[v.ID] = v
	}
	for _, v := range s.Projects {
		m.projects[v.ID] = v
	}
	for _, v := range s.Donations {
		m.donations[v.ID] = v
	}
	for _, v := range s.Ledgers {
		m.ledgers[v.ID] = v
	}
	for _, v := range s.Expenses {
		m.expenses[v.ID] = v
	}
	for _, v := range s.Progress {
		m.progress[v.ID] = v
	}
	for _, v := range s.Follows {
		m.follows[v.ID] = v
	}
	for _, v := range s.Comments {
		m.comments[v.ID] = v
	}
	for _, v := range s.Receipts {
		m.receipts[v.ID] = v
		m.receiptsByCode[v.Code] = v.ID
	}
	for _, v := range s.Notifications {
		m.notifications[v.ID] = v
	}
	for _, v := range s.Audits {
		m.audits[v.ID] = v
	}
}
