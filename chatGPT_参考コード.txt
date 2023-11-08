import React, { useState, useEffect } from 'react';

const App = () => {
  const [data, setData] = useState([]); // データのリスト
  const [selectedItemId, setSelectedItemId] = useState(null); // 選択されたアイテムのID
  const [newItem, setNewItem] = useState({ name: '', description: '' });

  // バックエンドからデータを取得
  useEffect(() => {
    fetch('/api/data') // バックエンドのAPIエンドポイントを指定
      .then(response => response.json())
      .then(data => {
        setData(data);
      })
      .catch(error => {
        console.error('データの取得に失敗しました: ', error);
      });
  }, []);

  const handleCreateItem = () => {
    // 新しいアイテムを作成
    fetch('/api/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(newItem),
    })
      .then(response => response.json())
      .then(createdItem => {
        setData([...data, createdItem]); // 新しいアイテムをリストに追加
        setNewItem({ name: '', description: '' }); // 入力フィールドをリセット
      })
      .catch(error => {
        console.error('アイテムの作成に失敗しました: ', error);
      });
  };

  const handleUpdateItem = () => {
    // 選択されたアイテムを更新
    const updatedData = { 
      id: selectedItemId, // 更新対象のアイテムのID
      name: newItem.name, // 新しい名前
      description: newItem.description, // 新しい説明
    };
    fetch(`/api/update/${selectedItemId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updatedData),
    })
      .then(response => response.json())
      .then(updatedItem => {
        const updatedData = data.map(item =>
          item.id === selectedItemId ? updatedItem : item
        );
        setData(updatedData);
        setSelectedItemId(null); // 編集モードを終了
      })
      .catch(error => {
        console.error('アイテムの更新に失敗しました: ', error);
      });
  };

  const handleDeleteItem = () => {
    // 選択されたアイテムを削除
    fetch(`/api/delete/${selectedItemId}`, {
      method: 'DELETE',
    })
      .then(() => {
        const updatedData = data.filter(item => item.id !== selectedItemId);
        setData(updatedData);
        setSelectedItemId(null); // 削除後、選択を解除
      })
      .catch(error => {
        console.error('アイテムの削除に失敗しました: ', error);
      });
  };

  return (
    <div>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {data.map(item => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td>{item.description}</td>
              <td>
                <button onClick={() => setSelectedItemId(item.id)}>Edit</button>
                <button onClick={() => handleDeleteItem(item.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {selectedItemId !== null ? (
        <div>
          <h2>Edit Item</h2>
          <input
            type="text"
            value={newItem.name}
            onChange={(e) => setNewItem({ ...newItem, name: e.target.value })}
          />
          <input
            type="text"
            value={newItem.description}
            onChange={(e) => setNewItem({ ...newItem, description: e.target.value })}
          />
          <button onClick={handleUpdateItem}>Save Changes</button>
        </div>
      ) : (
        <div>
          <h2>Create New Item</h2>
          <input
            type="text"
            value={newItem.name}
            onChange={(e) => setNewItem({ ...newItem, name: e.target.value })}
          />
          <input
            type="text"
            value={newItem.description}
            onChange={(e) => setNewItem({ ...newItem, description: e.target.value })}
          />
          <button onClick={handleCreateItem}>Create</button>
        </div>
      )}
    </div>
  );
};

export default App;
