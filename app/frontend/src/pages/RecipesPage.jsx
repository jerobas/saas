import { useState } from 'react';
import { motion } from 'framer-motion';
import { Plus, Trash, Copy } from 'phosphor-react';

const RecipesPage = () => {
  const items = [
    { id: 1, name: 'Farinha', unit: 'kg', costPerUnit: 5.0 },
    { id: 2, name: 'Açúcar', unit: 'kg', costPerUnit: 3.5 },
    { id: 3, name: 'Ovos', unit: 'un', costPerUnit: 0.5 },
  ];

  const recipes = [
    {
      id: 1,
      name: 'Bolo de Chocolate',
      ingredients: [
        { itemId: 1, itemName: 'Farinha', quantity: 2, unit: 'kg', costPerUnit: 5.0 },
        { itemId: 2, itemName: 'AçúCar', quantity: 1, unit: 'kg', costPerUnit: 3.5 },
        { itemId: 3, itemName: 'Ovos', quantity: 6, unit: 'un', costPerUnit: 0.5 },
      ],
    },
  ];

  const addRecipe = (recipe) => {
    console.log('Adding recipe:', recipe);
  };

  const removeRecipe = (id) => {
    console.log('Removing recipe with id:', id);
  };

  const updateRecipe = (id, updatedRecipe) => {
    console.log('Updating recipe with id:', id, updatedRecipe);
  };

  const [openDialog, setOpenDialog] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [newRecipe, setNewRecipe] = useState({ name: '', ingredients: [] });
  const [selectedItem, setSelectedItem] = useState(null);
  const [itemQuantity, setItemQuantity] = useState('');

  const calculateRecipeCost = (ingredients) => {
    return ingredients.reduce((total, item) => total + (item.quantity * item.costPerUnit), 0);
  };

  const handleAddIngredient = () => {
    if (selectedItem && itemQuantity) {
      const item = items.find(i => i.id === parseInt(selectedItem));
      const newIngredient = {
        itemId: item.id,
        itemName: item.name,
        quantity: parseFloat(itemQuantity),
        unit: item.unit,
        costPerUnit: item.costPerUnit,
      };
      setNewRecipe({
        ...newRecipe,
        ingredients: [...newRecipe.ingredients, newIngredient],
      });
      setSelectedItem(null);
      setItemQuantity('');
    }
  };

  const handleRemoveIngredient = (index) => {
    setNewRecipe({
      ...newRecipe,
      ingredients: newRecipe.ingredients.filter((_, i) => i !== index),
    });
  };

  const handleSaveRecipe = () => {
    if (newRecipe.name && newRecipe.ingredients.length > 0) {
      if (editingId) {
        updateRecipe(editingId, newRecipe);
        setEditingId(null);
      } else {
        addRecipe({
          name: newRecipe.name,
          ingredients: newRecipe.ingredients,
        });
      }
      setNewRecipe({ name: '', ingredients: [] });
      setOpenDialog(false);
    }
  };

  const handleEditRecipe = (recipe) => {
    setNewRecipe(recipe);
    setEditingId(recipe.id);
    setOpenDialog(true);
  };

  const handleDeleteRecipe = (id) => {
    removeRecipe(id);
  };

  return (
    <>
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Receitas</h1>
              <p className="text-slate-600 mt-2">Crie receitas combinando ingredientes</p>
            </div>
            <button
              onClick={() => {
                setEditingId(null);
                setNewRecipe({ name: '', ingredients: [] });
                setOpenDialog(true);
              }}
              className="flex items-center gap-2 bg-pink-600 text-white px-6 py-3 rounded-lg hover:bg-pink-700 transition-all"
            >
              <Plus size={20} />
              Nova Receita
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-6"
        >
          {recipes.map((recipe) => (
            <div key={recipe.id} className="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h2 className="text-2xl font-bold text-slate-900">{recipe.name}</h2>
                  <p className="text-slate-600 mt-1">Custo por render: R$ {calculateRecipeCost(recipe.ingredients).toFixed(2)}</p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleEditRecipe(recipe)}
                    className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-all"
                  >
                    <Copy size={20} />
                  </button>
                  <button
                    onClick={() => handleDeleteRecipe(recipe.id)}
                    className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  >
                    <Trash size={20} />
                  </button>
                </div>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">Ingrediente</th>
                      <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">Quantidade</th>
                      <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">Custo Unit.</th>
                      <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">Subtotal</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recipe.ingredients.map((ing, idx) => (
                      <tr key={idx} className="border-b border-slate-100">
                        <td className="px-4 py-3 text-sm text-slate-900">{ing.itemName}</td>
                        <td className="px-4 py-3 text-sm text-slate-600">
                          {ing.quantity.toFixed(2)} {ing.unit}
                        </td>
                        <td className="px-4 py-3 text-sm text-slate-600">
                          R$ {ing.costPerUnit.toFixed(2)}
                        </td>
                        <td className="px-4 py-3 text-sm font-semibold text-slate-900">
                          R$ {(ing.quantity * ing.costPerUnit).toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </motion.div>

        {/* Add Recipe Dialog */}
        {openDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-white rounded-2xl p-8 max-w-2xl w-full shadow-xl my-8"
            >
              <h2 className="text-2xl font-bold text-slate-900 mb-6">
                {editingId ? 'Editar Receita' : 'Nova Receita'}
              </h2>

              <div className="space-y-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Nome da Receita</label>
                  <input
                    type="text"
                    value={newRecipe.name}
                    onChange={(e) => setNewRecipe({ ...newRecipe, name: e.target.value })}
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="Ex: Bolo de Chocolate"
                  />
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-slate-900 mb-3">Ingredientes</h3>
                  <div className="space-y-3">
                  <div className="flex gap-2">
                      <select
                        value={selectedItem || ''}
                        onChange={(e) => setSelectedItem(e.target.value)}
                        className="flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                      >
                        <option value="">Selecione um item</option>
                        {items.map(item => (
                          <option key={item.id} value={item.id}>
                            {item.name} ({item.unit})
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        value={itemQuantity}
                        onChange={(e) => setItemQuantity(e.target.value)}
                        placeholder="Qtd"
                        className="w-24 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                      />
                      {selectedItem && (
                        <div className="flex items-center px-3 py-2 bg-pink-50 rounded-lg">
                          <span className="text-sm font-medium text-pink-600">
                            {inventoryItems.find(i => i.id === parseInt(selectedItem))?.unit}
                          </span>
                        </div>
                      )}
                      <button
                        onClick={handleAddIngredient}
                        className="px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all"
                      >
                        <Plus size={20} />
                      </button>
                    </div>

                    {newRecipe.ingredients.length > 0 && (
                      <div className="bg-slate-50 rounded-lg p-3 max-h-48 overflow-y-auto">
                        <div className="space-y-2">
                          {newRecipe.ingredients.map((ing, idx) => (
                            <div key={idx} className="flex justify-between items-center bg-white p-2 rounded border border-slate-200">
                              <span className="text-sm text-slate-900">
                                {ing.itemName}: {ing.quantity.toFixed(2)} {ing.unit}
                              </span>
                              <button
                                onClick={() => handleRemoveIngredient(idx)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Trash size={16} />
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {newRecipe.ingredients.length > 0 && (
                  <div className="bg-pink-50 rounded-lg p-4">
                    <p className="text-sm text-slate-600">Custo total da receita:</p>
                    <p className="text-2xl font-bold text-pink-600">
                      R$ {calculateRecipeCost(newRecipe.ingredients).toFixed(2)}
                    </p>
                  </div>
                )}
              </div>

              <div className="flex gap-4">
                <button
                  onClick={() => setOpenDialog(false)}
                  className="flex-1 px-4 py-2 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all"
                >
                  Cancelar
                </button>
                <button
                  onClick={handleSaveRecipe}
                  disabled={!newRecipe.name || newRecipe.ingredients.length === 0}
                  className="flex-1 px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300"
                >
                  Salvar
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </main>
    </>
  );
};

export default RecipesPage;
