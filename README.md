# Sistema de cache en Golang

Este proyecto utiliza como motor de base de datos [Postgres](https://www.postgresql.org/).  
Para el almacenamiento de cache se utiliza [Redis](https://redis.io/), en un contenedor de docker compose[Docker](https://docs.docker.com/compose/)
Para la comunicación entre los servicios se utiliza [gRPC](https://grpc.io/), con su implementación en [Go](https://grpc.io/docs/languages/go/basics/).  
Para la lectura de la información y su posterior almacenamiento en la base de datos se utilizan los comandos de postgres

Para la comunicación con Redis desde Go se utilizó [go-redis](https://redis.io/docs/latest/develop/connect/clients/go/) 
## Preparando el servicio

* Para crear la base de datos, esto se hará con el motor Postgres que debe ser instalado previamente en el ordenador personal, debido a conflictos con postgres y docker se decidio esto.
* En este caso se usara el siguiente [Archivo](https://drive.google.com/file/d/1I72j-FSfVAsjXySbOEYDVtZscuDxdMdg/view?usp=sharing).

* En donde se tiene que crear la siguiente tabla en la base de datos con los siguientes parametros: Database: postgres, Usuario: postgres, Password: 123, sslmode=disable.

```sql
create table public.Universitario(
	cat_periodo int,
	codigo_unico varchar(200),
	mrun int,
	gen_alu int,
	fec_nac_alu int,
	rango_edada varchar(200),
	anio_ing_carr_ori int,
	sem_ing_carr_ori int,
	anio_ing_carr_act int,
	sem_ing_carr_act int,
	nomb_titulo_obtenido varchar(200),
	nomb_grado_obtenido varchar(200),
	fecha_obtencion_titulo int,
	tipo_inst_1 varchar(200),
	tipo_inst_2 varchar(200),
	tipo_inst_3 varchar(200),
	cod_inst int,
	nomb_inst varchar(200),
	cod_sede int,
	nomb_sede varchar(200),
	cod_carrera int,
	nomb_carrera varchar(200),
	nivel_global varchar(200),
	nivel_carrera_1 varchar(200),
	nivel_carrera_2 varchar(200),
	dur_estudio_carr int, 
	dur_proceso_tit int,
	dur_total_carr int,
	region_sede varchar(200),
	provincia_sede varchar(200),
	comuna_sede varchar(200),
	jornada varchar(200),
	modalidad varchar(200),
	version int, 
	tipo_plan_carr  varchar(200),
	area_cineunesco  varchar(200),
	area_cine_f_97  varchar(200),
	subarea_cine_f_97  varchar(200),
	area_cine_f_13  varchar(200),
	subarea_cine_f_13  varchar(200),
	area_carrera_generica_n  varchar(200)

);
```
Luego para el ingreso de estos datos se hace mediante copiado de postgres
```sql
\COPY public.Universitario FROM 'PATH\Universitario.csv' delimiter ',' csv header;
```

* Compilar protoc
```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/grpc_cache.proto
```
## Iniciando docker-compose
Para el inicio de las instancias de redis hay que realizar:
```bash
docker-compose up --build --remove-orphans
```
## Comprobacion correcta de la base de datos
Para verificar si es que se puede ejecutar las consultas en la base de datos es necesario ejecutar la verificacion del ingreso a la base de datos:
```bash
go run main.go
```
## Iniciando cliente y servidor

* Iniciar servidor
```bash
go run server/main.go
```

* Iniciar cliente
```bash
go run client/main.go
```

## Usando el cliente

Una vez se ejecuta el cliente, iniciara la busqueda de cierto valor de nombre de carrera de cada dato de el dataset de universidades, que contiene el string nombre de carrera.

El cliente buscará en el cache de Redis, si no se encuentra, se buscará en la base de datos y se almacenará en el cache.

## Analizando cache

Para esta experiencia se tomo en cuenta los porcentajes de acierto que posee en relacion a los datos solicitados que estan en cache y los que no, en donde para todos los casos se limito la cache a 10mb para tener igualdad de condiciones
