create table ts_checks_m1_2016
    (check (created >= TIMESTAMP '2016-01-01 00:00:00-00' and created < TIMESTAMP '2016-02-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m2_2016
    (check (created >= TIMESTAMP '2016-02-01 00:00:00-00' and created < TIMESTAMP '2016-03-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m3_2016
    (check (created >= TIMESTAMP '2016-03-01 00:00:00-00' and created < TIMESTAMP '2016-04-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m4_2016
    (check (created >= TIMESTAMP '2016-04-01 00:00:00-00' and created < TIMESTAMP '2016-05-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m5_2016
    (check (created >= TIMESTAMP '2016-05-01 00:00:00-00' and created < TIMESTAMP '2016-06-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m6_2016
    (check (created >= TIMESTAMP '2016-06-01 00:00:00-00' and created < TIMESTAMP '2016-07-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m7_2016
    (check (created >= TIMESTAMP '2016-07-01 00:00:00-00' and created < TIMESTAMP '2016-08-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m8_2016
    (check (created >= TIMESTAMP '2016-08-01 00:00:00-00' and created < TIMESTAMP '2016-09-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m9_2016
    (check (created >= TIMESTAMP '2016-09-01 00:00:00-00' and created < TIMESTAMP '2016-10-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m10_2016
    (check (created >= TIMESTAMP '2016-10-01 00:00:00-00' and created < TIMESTAMP '2016-11-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m11_2016
    (check (created >= TIMESTAMP '2016-11-01 00:00:00-00' and created < TIMESTAMP '2016-12-01 00:00:00-00'))
    inherits (ts_checks);

create table ts_checks_m12_2016
    (check (created >= TIMESTAMP '2016-12-01 00:00:00-00' and created < TIMESTAMP '2017-01-01 00:00:00-00'))
    inherits (ts_checks);

create index idx_ts_checks_m1_2016_created on ts_checks_m1_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m2_2016_created on ts_checks_m2_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m3_2016_created on ts_checks_m3_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m4_2016_created on ts_checks_m4_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m5_2016_created on ts_checks_m5_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m6_2016_created on ts_checks_m6_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m7_2016_created on ts_checks_m7_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m8_2016_created on ts_checks_m8_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m9_2016_created on ts_checks_m9_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m10_2016_created on ts_checks_m10_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m11_2016_created on ts_checks_m11_2016 using brin (cluster_id, check_id, created);
create index idx_ts_checks_m12_2016_created on ts_checks_m12_2016 using brin (cluster_id, check_id, created);

create or replace function on_ts_checks_insert_2016() returns trigger as $$
begin
    if ( new.created >= TIMESTAMP '2016-01-01 00:00:00-00' and new.created < TIMESTAMP '2016-02-01 00:00:00-00') then
        insert into ts_checks_m1_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-02-01 00:00:00-00' and new.created < TIMESTAMP '2016-03-01 00:00:00-00') then
        insert into ts_checks_m2_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-03-01 00:00:00-00' and new.created < TIMESTAMP '2016-04-01 00:00:00-00') then
        insert into ts_checks_m3_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-04-01 00:00:00-00' and new.created < TIMESTAMP '2016-05-01 00:00:00-00') then
        insert into ts_checks_m4_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-05-01 00:00:00-00' and new.created < TIMESTAMP '2016-06-01 00:00:00-00') then
        insert into ts_checks_m5_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-06-01 00:00:00-00' and new.created < TIMESTAMP '2016-07-01 00:00:00-00') then
        insert into ts_checks_m6_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-07-01 00:00:00-00' and new.created < TIMESTAMP '2016-08-01 00:00:00-00') then
        insert into ts_checks_m7_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-08-01 00:00:00-00' and new.created < TIMESTAMP '2016-09-01 00:00:00-00') then
        insert into ts_checks_m8_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-09-01 00:00:00-00' and new.created < TIMESTAMP '2016-10-01 00:00:00-00') then
        insert into ts_checks_m9_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-10-01 00:00:00-00' and new.created < TIMESTAMP '2016-11-01 00:00:00-00') then
        insert into ts_checks_m10_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-11-01 00:00:00-00' and new.created < TIMESTAMP '2016-12-01 00:00:00-00') then
        insert into ts_checks_m11_2016 values (new.*);
    elsif ( new.created >= TIMESTAMP '2016-12-01 00:00:00-00' and new.created < TIMESTAMP '2017-01-01 00:00:00-00') then
        insert into ts_checks_m12_2016 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_checks_insert_2016
    before insert on ts_checks
    for each row execute procedure on_ts_checks_insert_2016();
