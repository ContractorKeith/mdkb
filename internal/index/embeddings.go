package index

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
)

type EmbedNote struct {
	ID, Title, Body string
	MTime, Size     int64
	Current         bool
}
type Vector struct {
	ID, Title, Body string
	Chunk           int
	Values          []float32
}
type EmbeddingCounts struct {
	Model  string `json:"model"`
	Notes  int    `json:"notes"`
	Chunks int    `json:"chunks"`
}

func (i *Index) EmbeddingCounts(ctx context.Context, model string) (EmbeddingCounts, error) {
	c := EmbeddingCounts{Model: model}
	err := i.db.QueryRowContext(ctx, "SELECT count(DISTINCT note_id),count(*) FROM embeddings WHERE model=?", model).Scan(&c.Notes, &c.Chunks)
	return c, Error(err)
}
func (i *Index) EmbedNotes(ctx context.Context, model string) ([]EmbedNote, error) {
	rows, err := i.db.QueryContext(ctx, `SELECT n.id,n.title,n.body,n.mtime,n.size,coalesce(s.model=? AND s.mtime=n.mtime AND s.size=n.size,0) FROM notes n LEFT JOIN embedding_state s ON s.note_id=n.id ORDER BY n.id`, model)
	if err != nil {
		return nil, Error(err)
	}
	defer rows.Close()
	result := []EmbedNote{}
	for rows.Next() {
		var n EmbedNote
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.MTime, &n.Size, &n.Current); err != nil {
			return nil, Error(err)
		}
		result = append(result, n)
	}
	return result, Error(rows.Err())
}

// StoreEmbeddings publishes a complete note only if its indexed stamp still matches.
func (i *Index) StoreEmbeddings(ctx context.Context, n EmbedNote, model string, vectors [][]float32) error {
	if len(vectors) == 0 {
		return Error(fmt.Errorf("empty embeddings"))
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return Error(err)
	}
	defer tx.Rollback()
	var valid int
	if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM notes WHERE id=? AND mtime=? AND size=?", n.ID, n.MTime, n.Size).Scan(&valid); err != nil {
		return Error(err)
	}
	if valid != 1 {
		return Error(fmt.Errorf("note changed while embedding: %s", n.ID))
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM embeddings WHERE note_id=?", n.ID); err != nil {
		return Error(err)
	}
	for chunk, v := range vectors {
		if len(v) == 0 || len(v) != len(vectors[0]) {
			return Error(fmt.Errorf("invalid vector dimensions"))
		}
		raw := make([]byte, len(v)*4)
		for j, x := range v {
			binary.LittleEndian.PutUint32(raw[j*4:], math.Float32bits(x))
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO embeddings VALUES(?,?,?,?,?)", n.ID, model, chunk, len(v), raw); err != nil {
			return Error(err)
		}
	}
	if _, err = tx.ExecContext(ctx, "INSERT OR REPLACE INTO embedding_state VALUES(?,?,?,?)", n.ID, model, n.MTime, n.Size); err != nil {
		return Error(err)
	}
	return Error(tx.Commit())
}
func (i *Index) Vectors(ctx context.Context, model string) ([]Vector, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT n.id,n.title,n.body,e.chunk,e.dim,e.vec FROM embeddings e JOIN notes n ON n.id=e.note_id WHERE e.model=? ORDER BY n.id,e.chunk", model)
	if err != nil {
		return nil, Error(err)
	}
	defer rows.Close()
	result := []Vector{}
	for rows.Next() {
		var v Vector
		var dim int
		var raw []byte
		if err := rows.Scan(&v.ID, &v.Title, &v.Body, &v.Chunk, &dim, &raw); err != nil {
			return nil, Error(err)
		}
		if dim < 1 || len(raw)/4 != dim || len(raw)%4 != 0 {
			return nil, Error(fmt.Errorf("invalid stored vector"))
		}
		v.Values = make([]float32, dim)
		for j := range v.Values {
			v.Values[j] = math.Float32frombits(binary.LittleEndian.Uint32(raw[j*4:]))
		}
		result = append(result, v)
	}
	return result, Error(rows.Err())
}
