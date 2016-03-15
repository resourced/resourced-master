create table ts_logs_m1_2016
    (check (created at time zone 'utc' >= date '2016-01-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-01-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m2_2016
    (check (created at time zone 'utc' >= date '2016-02-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-02-29' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m3_2016
    (check (created at time zone 'utc' >= date '2016-03-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-03-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m4_2016
    (check (created at time zone 'utc' >= date '2016-04-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-04-30' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m5_2016
    (check (created at time zone 'utc' >= date '2016-05-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-05-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m6_2016
    (check (created at time zone 'utc' >= date '2016-06-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-06-30' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m7_2016
    (check (created at time zone 'utc' >= date '2016-07-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-07-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m8_2016
    (check (created at time zone 'utc' >= date '2016-08-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-08-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m9_2016
    (check (created at time zone 'utc' >= date '2016-09-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-09-30' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m10_2016
    (check (created at time zone 'utc' >= date '2016-10-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-10-31' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m11_2016
    (check (created at time zone 'utc' >= date '2016-11-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-11-30' at time zone 'utc'))
    inherits (ts_logs);

create table ts_logs_m12_2016
    (check (created at time zone 'utc' >= date '2016-12-01' at time zone 'utc' and created at time zone 'utc' <= date '2016-12-31' at time zone 'utc'))
    inherits (ts_logs);

create index idx_ts_logs_m1_2016_search on ts_logs_m1_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m2_2016_search on ts_logs_m2_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m3_2016_search on ts_logs_m3_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m4_2016_search on ts_logs_m4_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m5_2016_search on ts_logs_m5_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m6_2016_search on ts_logs_m6_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m7_2016_search on ts_logs_m7_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m8_2016_search on ts_logs_m8_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m9_2016_search on ts_logs_m9_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m10_2016_search on ts_logs_m10_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m11_2016_search on ts_logs_m11_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));
create index idx_ts_logs_m12_2016_search on ts_logs_m12_2016 using gin(cluster_id, created, hostname, tags, to_tsvector('english', logline));

create or replace function on_ts_logs_insert_2016() returns trigger as $$
begin
    if ( new.created at time zone 'utc' >= date '2016-01-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-01-31' at time zone 'utc') then
        insert into ts_logs_m1_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-02-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-02-29' at time zone 'utc') then
        insert into ts_logs_m2_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-03-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-03-31' at time zone 'utc') then
        insert into ts_logs_m3_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-04-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-04-30' at time zone 'utc') then
        insert into ts_logs_m4_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-05-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-05-31' at time zone 'utc') then
        insert into ts_logs_m5_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-06-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-06-30' at time zone 'utc') then
        insert into ts_logs_m6_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-07-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-07-31' at time zone 'utc') then
        insert into ts_logs_m7_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-08-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-08-31' at time zone 'utc') then
        insert into ts_logs_m8_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-09-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-09-30' at time zone 'utc') then
        insert into ts_logs_m9_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-10-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-10-31' at time zone 'utc') then
        insert into ts_logs_m10_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-11-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-11-30' at time zone 'utc') then
        insert into ts_logs_m11_2016 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2016-12-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2016-12-31' at time zone 'utc') then
        insert into ts_logs_m12_2016 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_logs_insert_2016
    before insert on ts_logs
    for each row execute procedure on_ts_logs_insert_2016();
