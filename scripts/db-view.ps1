# Скрипт для просмотра БД Atlas

Write-Host "=== Atlas Database Viewer ===" -ForegroundColor Cyan
Write-Host ""

# Функция для выполнения SQL запросов
function Invoke-DbQuery {
    param([string]$query)
    docker exec atlas-postgres psql -U atlas -d atlas -c $query
}

# Меню
while ($true) {
    Write-Host ""
    Write-Host "Выберите действие:" -ForegroundColor Yellow
    Write-Host "1. Показать все пользователи"
    Write-Host "2. Показать все workspace"
    Write-Host "3. Показать все каналы"
    Write-Host "4. Показать все сообщения"
    Write-Host "5. Выполнить свой SQL запрос"
    Write-Host "6. Открыть Adminer в браузере"
    Write-Host "0. Выход"
    Write-Host ""
    
    $choice = Read-Host "Ваш выбор"
    
    switch ($choice) {
        "1" {
            Write-Host "`nПользователи:" -ForegroundColor Green
            Invoke-DbQuery "SELECT id, email, display_name, created_at FROM users ORDER BY created_at DESC;"
        }
        "2" {
            Write-Host "`nWorkspace:" -ForegroundColor Green
            Invoke-DbQuery "SELECT w.id, w.name, u.display_name as owner, w.created_at FROM workspaces w LEFT JOIN users u ON w.owner_id = u.id ORDER BY w.created_at DESC;"
        }
        "3" {
            Write-Host "`nКаналы:" -ForegroundColor Green
            Invoke-DbQuery "SELECT c.id, c.name, c.type, c.is_private, w.name as workspace, c.created_at FROM channels c LEFT JOIN workspaces w ON c.workspace_id = w.id ORDER BY c.created_at DESC;"
        }
        "4" {
            Write-Host "`nСообщения (последние 10):" -ForegroundColor Green
            Invoke-DbQuery "SELECT m.id, u.display_name as author, LEFT(m.content, 50) as content, m.created_at FROM messages m LEFT JOIN users u ON m.user_id = u.id ORDER BY m.created_at DESC LIMIT 10;"
        }
        "5" {
            $sql = Read-Host "`nВведите SQL запрос"
            Invoke-DbQuery $sql
        }
        "6" {
            Write-Host "`nОткрываем Adminer..." -ForegroundColor Green
            Start-Process "http://localhost:8081"
            Write-Host "Данные для входа:" -ForegroundColor Cyan
            Write-Host "  Сервер: atlas-postgres"
            Write-Host "  Пользователь: atlas"
            Write-Host "  Пароль: atlas"
            Write-Host "  База данных: atlas"
        }
        "0" {
            Write-Host "Выход..." -ForegroundColor Yellow
            exit
        }
        default {
            Write-Host "Неверный выбор!" -ForegroundColor Red
        }
    }
}
