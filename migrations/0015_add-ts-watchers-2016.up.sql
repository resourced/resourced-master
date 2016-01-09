create table ts_watchers_q1_2016
    (check (created >= date '2016-01-01' and created <= date '2016-03-31'))
    inherits (ts_watchers);

create table ts_watchers_q2_2016
    (check (created >= date '2016-04-01' and created <= date '2016-06-30'))
    inherits (ts_watchers);

create table ts_watchers_q3_2016
    (check (created >= date '2016-07-01' and created <= date '2016-09-30'))
    inherits (ts_watchers);

create table ts_watchers_q4_2016
    (check (created >= date '2016-10-01' and created <= date '2016-12-31'))
    inherits (ts_watchers);

create index ts_watchers_q1_2016_created on ts_watchers_q1_2016 using brin (cluster_id, watcher_id, created);
create index ts_watchers_q2_2016_created on ts_watchers_q2_2016 using brin (cluster_id, watcher_id, created);
create index ts_watchers_q3_2016_created on ts_watchers_q3_2016 using brin (cluster_id, watcher_id, created);
create index ts_watchers_q4_2016_created on ts_watchers_q4_2016 using brin (cluster_id, watcher_id, created);

create or replace function on_ts_watchers_insert_2016() returns trigger as $$
begin
    if ( new.created >= date '2016-01-01' and new.created <= date '2016-03-31') then
        insert into ts_watchers_q1_2016 values (new.*);
    elsif ( new.created >= date '2016-04-01' and new.created <= date '2016-06-30') then
        insert into ts_watchers_q2_2016 values (new.*);
    elsif ( new.created >= date '2016-07-01' and new.created <= date '2016-09-30') then
        insert into ts_watchers_q3_2016 values (new.*);
    elsif ( new.created >= date '2016-10-01' and new.created <= date '2016-12-31') then
        insert into ts_watchers_q4_2016 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_watchers_insert_2016
    before insert on ts_watchers
    for each row execute procedure on_ts_watchers_insert_2016();
