package dbsync

import (
	"context"
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	changeLogTableDel = `DROP TABLE IF EXISTS change_log`
	changeLogTableDef = `
		CREATE TABLE change_log (
			id BIGSERIAL PRIMARY KEY,
			op TEXT NOT NULL,
			table_name TEXT NOT NULL,
			row_id TEXT NOT NULL,
			tx_id BIGINT,
			cmd_id cid,
			row_data JSONB
		)`

	changeLogFuncDel = `DROP FUNCTION IF EXISTS fn_process_change_log()`
	changeLogFuncDef = `
		CREATE OR REPLACE FUNCTION fn_process_change_log() RETURNS TRIGGER AS $$
		DECLARE
			cur_state enum_switchover_state := 'idle';
		BEGIN
			SELECT INTO cur_state current_state
			FROM switchover_state;
			
			IF cur_state != 'in_progress' THEN
				RETURN NEW;
			END IF;
		
			IF (TG_OP = 'DELETE') THEN
				INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id)
				VALUES (TG_OP, TG_TABLE_NAME, cast(OLD.id as TEXT), txid_current(), OLD.cmax);
				RETURN OLD;
			ELSE
				INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id, row_data)
				VALUES (TG_OP, TG_TABLE_NAME, cast(NEW.id as TEXT), txid_current(), NEW.cmin, to_jsonb(NEW));
				RETURN NEW;
			END IF;
		
			RETURN NULL;
		END;
		$$ LANGUAGE 'plpgsql'`
)

func changeLogTrigName(tableName string) string {
	return fmt.Sprintf("zz_99_change_log_%s", tableName)
}

func changeLogTrigDel(tableName string) string {
	return fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON %s`, pq.QuoteIdentifier(changeLogTrigName(tableName)), pq.QuoteIdentifier(tableName))
}
func changeLogTrigDef(tableName string) string {
	return fmt.Sprintf(`
		CREATE TRIGGER %s
		AFTER INSERT OR UPDATE OR DELETE ON %s
		FOR EACH ROW EXECUTE PROCEDURE fn_process_change_log()`,
		pq.QuoteIdentifier(changeLogTrigName(tableName)), pq.QuoteIdentifier(tableName))
}

func noTrigger(name string) bool {
	ignored := []string{"gorp_migrations", "change_log", "switchover_state"}
	for _, i := range ignored {
		if name == i {
			return true
		}
	}
	return false
}

// ChangeLogEnable will instrument the database for the sync operation.
func (s *Sync) ChangeLogEnable(ctx context.Context, sh *ishell.Context) error {
	var err error
	run := func(name, stmt string) {
		if err != nil {
			return
		}
		_, err = s.oldDB.ExecContext(ctx, stmt)
		err = errors.Wrap(err, name)
	}
	runNew := func(name, stmt string) {
		if err != nil {
			return
		}
		_, err = s.newDB.ExecContext(ctx, stmt)
		err = errors.Wrap(err, name)
	}
	sh.Println("Resetting change log...")
	runNew("configure dest change_log", changeLogTableDef)
	run("clear change_log", changeLogTableDel)
	run("configure change_log", changeLogTableDef)
	run("define change hook", changeLogFuncDef)

	sh.Println("Instrumenting tables...")
	for _, t := range s.tables {
		if noTrigger(t.Name) {
			continue
		}

		run("clear prev. trigger for "+t.SafeName(), changeLogTrigDel(t.Name))
		run("set trigger for "+t.SafeName(), changeLogTrigDef(t.Name))
	}
	if err != nil {
		return err
	}

	sh.Println("Activating change tracking...")
	res, err := s.oldDB.ExecContext(ctx, `update switchover_state set current_state = 'in_progress' where current_state = 'idle'`)
	if err != nil {
		return err
	}
	r, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if r != 1 {
		return errors.New("not idle")
	}

	return nil
}

// ChangeLogDisable will remove all sync instrumentation.
func (s *Sync) ChangeLogDisable(ctx context.Context, sh *ishell.Context) error {
	res, err := s.oldDB.ExecContext(ctx, `update switchover_state set current_state = 'idle' where current_state = 'in_progress'`)
	if err != nil {
		return err
	}
	r, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if r != 1 {
		return errors.New("not in_progress")
	}

	run := func(name, stmt string) {
		if err != nil {
			return
		}
		_, err = s.oldDB.ExecContext(ctx, stmt)
		err = errors.Wrap(err, name)
	}
	sh.Println("Removing triggers...")
	for _, t := range s.tables {
		run("clear trigger for "+t.SafeName(), changeLogTrigDel(t.Name))
	}

	sh.Println("Resetting change log...")
	run("remove change hook", changeLogFuncDel)
	run("remove change_log", changeLogTableDel)
	if err != nil {
		return err
	}

	return nil
}
