drop table dayticks;
create table dayticks (
	stockcode char(6) not null,
	date date,
	open float(6,2),
	high float(6,2),
	low float(6,2),
	close float(6,2));
