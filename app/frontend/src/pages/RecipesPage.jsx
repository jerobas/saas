/* eslint-disable no-unused-vars */
import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { Plus, Trash, Pencil, Package } from "phosphor-react";
import { GetAllItems } from "../../wailsjs/go/service/ItemService";
import {
  CreateRecipe,
  GetAllRecipes,
  UpdateRecipe,
  DeleteRecipe,
} from "../../wailsjs/go/service/RecipeService";

const RecipesPage = () => {
  const [items, setItems] = useState([]);
  const [recipes, setRecipes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);
  const [editingId, setEditingId] = useState(null);
  const [newRecipe, setNewRecipe] = useState({ name: "", ingredients: [] });
  const [selectedItem, setSelectedItem] = useState("");
  const [itemQuantity, setItemQuantity] = useState("");

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [itemsData, recipesData] = await Promise.all([
        GetAllItems(),
        GetAllRecipes(),
      ]);
      setItems(itemsData || []);
      setRecipes(recipesData || []);
    } catch (err) {
      console.error("Erro ao carregar dados:", err);
      setError("Erro ao carregar dados. Tente novamente.");
    } finally {
      setLoading(false);
    }
  };

  const calculateRecipeCost = (ingredients) => {
    return ingredients.reduce((total, ing) => {
      const item = items.find((i) => i.id === ing.item_id);
      if (!item) return total;
      // Aqui você pode buscar o preço médio dos batches se quiser
      // Por enquanto vamos usar 0 como placeholder
      return total + ing.quantity * 0;
    }, 0);
  };

  const getItemById = (itemId) => {
    return items.find((i) => i.id === itemId);
  };

  const handleAddIngredient = () => {
    if (selectedItem && itemQuantity) {
      const item = items.find((i) => i.id === selectedItem);
      if (!item) return;

      const newIngredient = {
        item_id: item.id,
        quantity: parseFloat(itemQuantity),
      };

      setNewRecipe({
        ...newRecipe,
        ingredients: [...newRecipe.ingredients, newIngredient],
      });
      setSelectedItem("");
      setItemQuantity("");
    }
  };

  const handleRemoveIngredient = (index) => {
    setNewRecipe({
      ...newRecipe,
      ingredients: newRecipe.ingredients.filter((_, i) => i !== index),
    });
  };

  const handleSaveRecipe = async () => {
    if (isSubmitting) return;

    if (newRecipe.name && newRecipe.ingredients.length > 0) {
      setIsSubmitting(true);
      setError(null);

      try {
        if (editingId) {
          // Atualizar receita existente
          await UpdateRecipe(editingId, newRecipe.name, newRecipe.ingredients);
        } else {
          // Criar nova receita
          await CreateRecipe(newRecipe.name, newRecipe.ingredients);
        }

        setNewRecipe({ name: "", ingredients: [] });
        setEditingId(null);
        setOpenDialog(false);
        await loadData();
      } catch (err) {
        console.error("Erro ao salvar receita:", err);
        setError("Erro ao salvar receita. Tente novamente.");
      } finally {
        setIsSubmitting(false);
      }
    }
  };

  const handleEditRecipe = (recipe) => {
    setNewRecipe({
      name: recipe.name,
      ingredients: recipe.ingredients || [],
    });
    setEditingId(recipe.id);
    setOpenDialog(true);
  };

  const handleDeleteRecipe = (id) => {
    if (loading) return;
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (deleteConfirmId && !loading) {
      try {
        setError(null);
        await DeleteRecipe(deleteConfirmId);
        setDeleteConfirmId(null);
        await loadData();
      } catch (err) {
        console.error("Erro ao deletar receita:", err);
        setError("Erro ao deletar receita. Tente novamente.");
      }
    }
  };

  return (
    <>
      {/* Loading Overlay */}
      {(loading || isSubmitting) && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center z-[60] backdrop-blur-md"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4"
          >
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
              className="w-12 h-12 border-4 border-pink-200 border-t-pink-600 rounded-full"
            />
            <div className="text-center">
              <h3 className="text-lg font-semibold text-slate-900">
                {isSubmitting
                  ? "Salvando receita..."
                  : "Carregando receitas..."}
              </h3>
              <p className="text-sm text-slate-600 mt-1">Por favor aguarde</p>
            </div>
          </motion.div>
        </motion.div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirmId && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center z-50 backdrop-blur-md"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0, y: 20 }}
            animate={{ scale: 1, opacity: 1, y: 0 }}
            className="bg-white rounded-2xl p-8 max-w-md w-full shadow-xl"
          >
            <div className="flex items-center justify-center w-12 h-12 bg-red-100 rounded-full mx-auto mb-6">
              <Trash size={24} className="text-red-600" />
            </div>

            <h2 className="text-2xl font-bold text-slate-900 mb-2 text-center">
              Deletar Receita?
            </h2>

            <p className="text-slate-600 text-center mb-6">
              Tem certeza que deseja remover esta receita? Esta ação não pode
              ser desfeita.
            </p>

            <div className="flex gap-4">
              <button
                onClick={() => setDeleteConfirmId(null)}
                disabled={loading}
                className="flex-1 px-4 py-3 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all disabled:bg-slate-50 disabled:cursor-not-allowed font-medium"
              >
                Cancelar
              </button>
              <button
                onClick={confirmDelete}
                disabled={loading}
                className="flex-1 px-4 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed font-medium flex items-center justify-center gap-2"
              >
                {loading ? (
                  <>
                    <motion.div
                      animate={{ rotate: 360 }}
                      transition={{
                        duration: 1,
                        repeat: Infinity,
                        ease: "linear",
                      }}
                      className="w-4 h-4 border-2 border-white border-t-transparent rounded-full"
                    />
                    Deletando...
                  </>
                ) : (
                  "Sim, Deletar"
                )}
              </button>
            </div>
          </motion.div>
        </motion.div>
      )}

      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Receitas</h1>
              <p className="text-slate-600 mt-2">
                Crie receitas combinando ingredientes
              </p>
            </div>
            <button
              onClick={() => {
                setEditingId(null);
                setNewRecipe({ name: "", ingredients: [] });
                setOpenDialog(true);
              }}
              disabled={loading || isSubmitting || items.length === 0}
              className="flex items-center gap-2 bg-pink-600 text-white px-6 py-3 rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Nova Receita
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* Error Message */}
        {error && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm mb-6"
          >
            <span>{error}</span>
            <button
              onClick={() => setError(null)}
              className="ml-auto text-red-600 hover:text-red-800"
            >
              ✕
            </button>
          </motion.div>
        )}

        {/* Summary Card */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm mb-8"
        >
          <div className="flex items-center gap-3">
            <Package size={24} className="text-pink-600" />
            <div>
              <h3 className="text-slate-600 text-sm font-medium">
                Total de Receitas
              </h3>
              <p className="text-3xl font-bold text-slate-900 mt-1">
                {recipes.length}
              </p>
            </div>
          </div>
        </motion.div>

        {/* Recipes List */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-6"
        >
          {recipes.length === 0 ? (
            <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-12 text-center">
              <Package size={48} className="mx-auto text-slate-300 mb-4" />
              <p className="text-slate-500">
                {items.length === 0
                  ? "Cadastre ingredientes primeiro para criar receitas."
                  : 'Nenhuma receita cadastrada. Clique em "Nova Receita" para começar.'}
              </p>
            </div>
          ) : (
            recipes.map((recipe) => (
              <div
                key={recipe.id}
                className="bg-white rounded-2xl border border-slate-100 shadow-sm p-6"
              >
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h2 className="text-2xl font-bold text-slate-900">
                      {recipe.name}
                    </h2>
                    <p className="text-slate-600 mt-1">
                      {recipe.ingredients?.length || 0} ingrediente(s)
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleEditRecipe(recipe)}
                      disabled={loading || isSubmitting}
                      className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-all disabled:text-slate-300 disabled:cursor-not-allowed"
                      title="Editar receita"
                    >
                      <Pencil size={20} />
                    </button>
                    <button
                      onClick={() => handleDeleteRecipe(recipe.id)}
                      disabled={loading || isSubmitting}
                      className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-all disabled:text-slate-300 disabled:cursor-not-allowed"
                      title="Deletar receita"
                    >
                      <Trash size={20} />
                    </button>
                  </div>
                </div>

                {recipe.ingredients && recipe.ingredients.length > 0 && (
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-slate-200">
                          <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">
                            Ingrediente
                          </th>
                          <th className="px-4 py-3 text-left text-sm font-semibold text-slate-900">
                            Quantidade
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {recipe.ingredients.map((ing, idx) => {
                          const item = getItemById(ing.item_id);
                          return (
                            <tr key={idx} className="border-b border-slate-100">
                              <td className="px-4 py-3 text-sm text-slate-900">
                                {item ? item.name : "Item não encontrado"}
                              </td>
                              <td className="px-4 py-3 text-sm text-slate-600">
                                {parseFloat(ing.quantity).toFixed(3)}{" "}
                                {item ? item.unit : ""}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ))
          )}
        </motion.div>
        {openDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto p-4">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-white rounded-2xl p-8 max-w-2xl w-full shadow-xl my-8"
            >
              <h2 className="text-2xl font-bold text-slate-900 mb-6">
                {editingId ? "Editar Receita" : "Nova Receita"}
              </h2>

              {error && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm mb-6"
                >
                  <span>{error}</span>
                </motion.div>
              )}

              <div className="space-y-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Nome da Receita
                  </label>
                  <input
                    type="text"
                    value={newRecipe.name}
                    onChange={(e) =>
                      setNewRecipe({ ...newRecipe, name: e.target.value })
                    }
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                    placeholder="Ex: Bolo de Chocolate"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-slate-900 mb-3">
                    Ingredientes
                  </h3>
                  <div className="space-y-3">
                    <div className="flex gap-2">
                      <select
                        value={selectedItem}
                        onChange={(e) => setSelectedItem(e.target.value)}
                        className="flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                        disabled={isSubmitting}
                      >
                        <option value="">Selecione um ingrediente</option>
                        {items.map((item) => (
                          <option key={item.id} value={item.id}>
                            {item.name} ({item.unit})
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        step="0.001"
                        value={itemQuantity}
                        onChange={(e) => setItemQuantity(e.target.value)}
                        placeholder="Quantidade"
                        className="w-32 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
                        disabled={isSubmitting}
                      />
                      {selectedItem && (
                        <div className="flex items-center px-3 py-2 bg-pink-50 rounded-lg">
                          <span className="text-sm font-medium text-pink-600">
                            {items.find((i) => i.id === selectedItem)?.unit}
                          </span>
                        </div>
                      )}
                      <button
                        onClick={handleAddIngredient}
                        disabled={
                          !selectedItem || !itemQuantity || isSubmitting
                        }
                        className="px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed"
                      >
                        <Plus size={20} />
                      </button>
                    </div>

                    {newRecipe.ingredients.length > 0 && (
                      <div className="bg-slate-50 rounded-lg p-3 max-h-64 overflow-y-auto">
                        <div className="space-y-2">
                          {newRecipe.ingredients.map((ing, idx) => {
                            const item = getItemById(ing.item_id);
                            return (
                              <div
                                key={idx}
                                className="flex justify-between items-center bg-white p-3 rounded border border-slate-200"
                              >
                                <span className="text-sm text-slate-900">
                                  <span className="font-medium">
                                    {item?.name || "Desconhecido"}
                                  </span>
                                  {": "}
                                  {parseFloat(ing.quantity).toFixed(3)}{" "}
                                  {item?.unit || ""}
                                </span>
                                <button
                                  onClick={() => handleRemoveIngredient(idx)}
                                  disabled={isSubmitting}
                                  className="text-red-600 hover:text-red-700 disabled:text-slate-300 disabled:cursor-not-allowed transition-colors"
                                >
                                  <Trash size={16} />
                                </button>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    )}

                    {newRecipe.ingredients.length === 0 && (
                      <div className="text-center py-8 text-slate-500 text-sm">
                        Adicione ingredientes à receita
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex gap-4">
                <button
                  onClick={() => {
                    setOpenDialog(false);
                    setError(null);
                  }}
                  disabled={isSubmitting}
                  className="flex-1 px-4 py-2 border border-slate-300 rounded-lg text-slate-700 hover:bg-slate-50 transition-all disabled:bg-slate-50 disabled:cursor-not-allowed"
                >
                  Cancelar
                </button>
                <button
                  onClick={handleSaveRecipe}
                  disabled={
                    isSubmitting ||
                    !newRecipe.name.trim() ||
                    newRecipe.ingredients.length === 0
                  }
                  className="flex-1 px-4 py-2 bg-pink-600 text-white rounded-lg hover:bg-pink-700 transition-all disabled:bg-slate-300 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isSubmitting ? (
                    <>
                      <motion.div
                        animate={{ rotate: 360 }}
                        transition={{
                          duration: 1,
                          repeat: Infinity,
                          ease: "linear",
                        }}
                        className="w-4 h-4 border-2 border-white border-t-transparent rounded-full"
                      />
                      Salvando...
                    </>
                  ) : (
                    "Salvar"
                  )}
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
