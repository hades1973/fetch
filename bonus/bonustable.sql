drop table bonus;
create table bonus (
	bonusid int unsigned not null auto_increment primary key,
	stockcode char(6) not null,
	Year year(4),
	GuJi float(4.1),
	SonGu float(4.1),
	ZengGu float(4.1),
	PaiXi float(5.2),
	GQDJR date,
	GQJZR date,
	HGSSR date);
