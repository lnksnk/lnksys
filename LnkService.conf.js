DBMSRegister("lnks", "driver=postgres", "username=lnksys", "password=lnksyslnksys", "host=127.0.0.1:5432", "sslmode=disable");
InvokeListener("0.0.0.0:1111");
MAPRoots("/","./","movies/","D:/movies/");