package webserver

const MainPage string = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WiFi Setup</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 400px;
            margin: 50px auto;
            padding: 20px;
        }
        
        .form-group {
            margin-bottom: 15px;
        }
        
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        
        input[type="text"],
        input[type="password"] {
            width: 100%;
            padding: 8px;
            border: 1px solid #ccc;
            border-radius: 4px;
            box-sizing: border-box;
        }
        
        button {
            background-color: #4CAF50;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        
        button:hover {
            background-color: #45a049;
        }
        
        .error {
            color: red;
            margin-top: 10px;
            display: none;
        }
        
        .success {
            color: green;
            margin-top: 10px;
            display: none;
        }
    </style>
</head>
<body>
    <h2>ПОдключиться к WiFi</h2>
    
    <form id="wifiForm">
        <div class="form-group">
            <label for="ssid">SSID:</label>
            <input type="text" id="ssid" name="ssid" required>
        </div>
        
        <div class="form-group">
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        
        <button type="submit">Отправить</button>
    </form>
    
    <div id="errorMsg" class="error">Пароль должен содержать минимум 8 символов!</div>
    <div id="successMsg" class="success">Настройки успешно отправлены!</div>

    <script>
        document.getElementById('wifiForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const ssid = document.getElementById('ssid').value;
            const password = document.getElementById('password').value;
            const errorMsg = document.getElementById('errorMsg');
            const successMsg = document.getElementById('successMsg');
            
            // Скрываем предыдущие сообщения
            errorMsg.style.display = 'none';
            successMsg.style.display = 'none';
            
            // Проверяем длину пароля
            if (password.length < 8) {
                errorMsg.style.display = 'block';
                return;
            }
            
            // Отправляем POST запрос на сервер
            fetch('/api/wifi/setup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ssid: ssid,
                    password: password
                })
            })
            .then(response => {
                if (response.ok) {
                    successMsg.style.display = 'block';
                    // Очищаем форму
                    document.getElementById('wifiForm').reset();
                } else {
                    throw new Error('Ошибка сервера');
                }
            })
            .catch(error => {
                errorMsg.textContent = 'Ошибка при отправке: ' + error.message;
                errorMsg.style.display = 'block';
            });
        });
    </script>
</body>
</html>`
