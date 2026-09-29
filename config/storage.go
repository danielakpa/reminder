package config
// Storage selection manager
//
// 1. Check current storage option
// 2. If storage is JSON:
//      - Load JSON files
//      - Use file storage
//
// 3. If storage is PostgreSQL:
//      - Connect database
//      - Use SQL queries
//
// 4. Return selected storage system