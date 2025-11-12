import os
import json
import time
import numpy as np
import faiss
from tqdm import tqdm
from sentence_transformers import SentenceTransformer
import gensim.downloader as api


# --- CARREGAR TEXTOS ---
def load_texts(folder):
    files, texts = [], []
    for f in os.listdir(folder):
        if f.endswith(".txt"):
            with open(os.path.join(folder, f), encoding="utf-8") as fp:
                texts.append(fp.read())
                files.append(f)
    return files, texts


# --- EMBEDDING: BERT MULTILÍNGUE ---
def embed_bert(texts, model):
    embs = model.encode(texts, convert_to_numpy=True)
    faiss.normalize_L2(embs)
    return embs.astype(np.float32)

# --- EMBEDDING: GLOVE ---
class GloveEmbedder:
    def __init__(self, model_name="glove-wiki-gigaword-300"):
        self.model = api.load(model_name)
        self.dim = self.model.vector_size

    def embed(self, texts):
        embs = []
        for t in texts:
            words = [w for w in t.split() if w in self.model]
            if words:
                vec = np.mean([self.model[w] for w in words], axis=0)
            else:
                vec = np.zeros(self.dim, dtype=np.float32)
            embs.append(vec)
        return np.array(embs, dtype=np.float32)

# --- EMBEDDING: WORD2VEC ---
class Word2VecEmbedder:
    def __init__(self, model_name="word2vec-google-news-300"):
        self.model = api.load(model_name)
        self.dim = self.model.vector_size

    def embed(self, texts):
        embs = []
        for t in texts:
            words = [w for w in t.split() if w in self.model]
            if words:
                vec = np.mean([self.model[w] for w in words], axis=0)
            else:
                vec = np.zeros(self.dim, dtype=np.float32)
            embs.append(vec)
        return np.array(embs, dtype=np.float32)


# --- FAISS (cosine) ---
def build_index(embeddings):
    dim = embeddings.shape[1]
    index = faiss.IndexFlatIP(dim)
    index.add(embeddings)
    return index


def search(index, query_emb, n_docs):
    query_emb = np.array(query_emb, dtype=np.float32)
    faiss.normalize_L2(query_emb)
    start = time.time()
    scores, idx = index.search(query_emb, n_docs)
    elapsed = int((time.time() - start) * 1_000_000)
    return (idx[0]).astype(np.uint16), elapsed  # base 1

# --- LOAD INPUTS ---
def load_inputs(json_path):
    with open(json_path, encoding="utf-8") as f:
        return json.load(f)


# --- MAIN ---
laws_folder = "misc/corpus/clean"
inputs_path = "misc/searchLegalInputs.json"

files, texts = load_texts(laws_folder)
n_docs = len(files)
TOP_K = None

print(f"{n_docs} arquivos carregados.")
print("Gerando embeddings dos documentos...")

# Modelo BERT multilíngue (melhor pra português)
bert_model = SentenceTransformer("paraphrase-multilingual-MiniLM-L12-v2")
embeddings_bert = embed_bert(texts, bert_model)
index_bert = build_index(embeddings_bert)

# Word2Vec
w2v_model = Word2VecEmbedder()
embeddings_w2v = w2v_model.embed(texts)
index_w2v = build_index(embeddings_w2v)

# GloVe
glove_model = GloveEmbedder()
embeddings_glove = glove_model.embed(texts)
index_glove = build_index(embeddings_glove)

# Inputs
inputs = load_inputs(inputs_path)
final_output = {"words10": [], "words20": [], "words40": []}

# --- GERA SAMPLE ---
def generate_sample(sample):
    text = sample["input"]
    query_w2v = w2v_model.embed([text])
    query_bert = bert_model.encode([text], convert_to_tensor=False)
    query_glove = glove_model.embed([text])


    ordered, t = search(index_glove, query_glove, n_docs)
    sample["glove"] = (ordered[:TOP_K] + 1).tolist() if TOP_K is not None else (ordered + 1).tolist()
    sample["gloveT"] = t

    ordered, t = search(index_bert, query_bert, n_docs)
    sample["bert"] = (ordered[:TOP_K] + 1).tolist() if TOP_K is not None else (ordered + 1).tolist()
    sample["bertT"] = t

    ordered, t = search(index_w2v, query_w2v, n_docs)
    sample["word2vec"] = (ordered[:TOP_K] + 1).tolist() if TOP_K is not None else (ordered + 1).tolist()
    sample["word2vecT"] = t

    return sample


# --- EXECUÇÃO ---
for group in ["words10", "words20", "words40"]:
    print(f"Processando grupo {group}...")
    for phrase in tqdm(inputs[group]):
        final_output[group].append(generate_sample(phrase))

with open("output.json", "w", encoding="utf-8") as f:
    json.dump(final_output, f, ensure_ascii=False, indent=2)

print("Finalizado! JSON salvo como output.json")
