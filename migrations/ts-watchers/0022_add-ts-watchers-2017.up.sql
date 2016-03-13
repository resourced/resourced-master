create table ts_watchers_m1_2017
    (check (created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-01-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m2_2017
    (check (created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-02-28' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m3_2017
    (check (created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-03-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m4_2017
    (check (created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-04-30' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m5_2017
    (check (created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-05-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m6_2017
    (check (created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-06-30' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m7_2017
    (check (created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-07-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m8_2017
    (check (created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-08-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m9_2017
    (check (created at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-09-30' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m10_2017
    (check (created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-10-31' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m11_2017
    (check (created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-11-30' at time zone 'utc'))
    inherits (ts_watchers);

create table ts_watchers_m12_2017
    (check (created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-12-31' at time zone 'utc'))
    inherits (ts_watchers);

create index idx_ts_watchers_m1_2017_created on ts_watchers_m1_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m2_2017_created on ts_watchers_m2_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m3_2017_created on ts_watchers_m3_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m4_2017_created on ts_watchers_m4_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m5_2017_created on ts_watchers_m5_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m6_2017_created on ts_watchers_m6_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m7_2017_created on ts_watchers_m7_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m8_2017_created on ts_watchers_m8_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m9_2017_created on ts_watchers_m9_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m10_2017_created on ts_watchers_m10_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m11_2017_created on ts_watchers_m11_2017 using brin (cluster_id, watcher_id, created);
create index idx_ts_watchers_m12_2017_created on ts_watchers_m12_2017 using brin (cluster_id, watcher_id, created);

create or replace function on_ts_watchers_insert_2017() returns trigger as $$
begin
    if ( new.created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-01-31' at time zone 'utc') then
        insert into ts_watchers_m1_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-02-28' at time zone 'utc') then
        insert into ts_watchers_m2_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-03-31' at time zone 'utc') then
        insert into ts_watchers_m3_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-04-30' at time zone 'utc') then
        insert into ts_watchers_m4_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-05-31' at time zone 'utc') then
        insert into ts_watchers_m5_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-06-30' at time zone 'utc') then
        insert into ts_watchers_m6_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-07-31' at time zone 'utc') then
        insert into ts_watchers_m7_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-08-31' at time zone 'utc') then
        insert into ts_watchers_m8_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-09-01' at time zone 'utc'and new.created at time zone 'utc' <= date '2017-09-30' at time zone 'utc') then
        insert into ts_watchers_m9_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-10-31' at time zone 'utc') then
        insert into ts_watchers_m10_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-11-30' at time zone 'utc') then
        insert into ts_watchers_m11_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-12-31' at time zone 'utc') then
        insert into ts_watchers_m12_2017 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_watchers_insert_2017
    before insert on ts_watchers
    for each row execute procedure on_ts_watchers_insert_2017();
