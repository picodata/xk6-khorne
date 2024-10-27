# xk6-khorne

Плагин, позволяющий запускать хаос-тесты на k8s	кластере при помощи chaos-mesh.


## Установка

Установите xk6
```
go install go.k6.io/xk6/cmd/xk6@latest
```

Соберите k6 с модулем xk6-khorne:
```bash
xk6 build --with github.com/picodata/xk6-khorne@latest
```

## Перед запуском тестов

До запуска, убедитесь, что у вас есть доступ к K8s кластеру, в котором установлен [chaosmesh](https://chaos-mesh.org/docs/production-installation-using-helm/). Тестируемый кластер и контроллер chaosmesh **должны находится в одном namespace**.

## Примеры использования
Примеры использования можно найти в папке ```/examples```

