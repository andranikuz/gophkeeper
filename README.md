# gophkeeper
Менеджер паролей GophKeeper

Команда генерации клиента и сервера
```shell
sh scripts/build.sh 
```

Запуск сервера на примере версии для Mac
```shell
./build/gophkeeper-server-darwin 
```
Хелпер для клиента
```shell
./build/gophkeeper-client-darwin
```
Регистрация пользователя на клиенте
```shell
./build/gophkeeper-client-darwin register -username=username -password=12345678
```
Аутентификация
```shell
./build/gophkeeper-client-darwin login -username=username -password=12345678
```
Сохранение текстовой информации
```shell
./build/gophkeeper-client-darwin save-text -text=nnnnnnnnnnnnn -meta=meta
```
Сохранение файла. Передаем абсолютный путь
```shell
./build/gophkeeper-client-darwin save-file -file=/Users/andranikuz/self/gophkeeper/README.md
```
Получение информации из локального хранилища
```shell
./build/gophkeeper-client-darwin get
```
Синхронизация с сервером
```shell
./build/gophkeeper-client-darwin sync
```