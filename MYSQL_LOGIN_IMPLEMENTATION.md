# MySQL Setup Guide

## Configure .env

```env
DB_TYPE=mysql
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=expensevault_app
MYSQL_PASSWORD=strong_password
MYSQL_DATABASE=expensevault
```

## Create database and user

```sql
CREATE DATABASE IF NOT EXISTS expensevault;
CREATE USER 'expensevault_app'@'%' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON expensevault.* TO 'expensevault_app'@'%';
FLUSH PRIVILEGES;
```

## Run the app

```bash
./expensevault tui
```

The app creates `transactions` and `users` tables automatically.
