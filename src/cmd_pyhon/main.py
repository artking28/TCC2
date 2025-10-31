import os
import json
import time
import numpy as np
import faiss
import openai
from tqdm import tqdm

# Transformers
from sentence_transformers import SentenceTransformer
from transformers import AutoTokenizer, AutoModel
import torch

# Word2Vec
import gensim.downloader as api

openai.api_key = "SUA_API_KEY"

# --- CARREGAR TEXTOS JURÍDICOS ---
def load_texts(folder):
    files, texts = [], []
    for f in os.listdir(folder):
        if f.endswith(".txt"):
            with open(os.path.join(folder, f), encoding="utf-8") as fp:
                texts.append(fp.read())
                files.append(f)
    return files, texts

# --- FUNÇÃO PARA EMBEDDINGS OPENAI ---
# def embed_openai(texts):
#     embs = []
#     for t in texts:
#         r = openai.embeddings.create(model="text-embedding-3-large", input=t)
#         embs.append(r.data[0].embedding)
#     return np.array(embs, dtype=np.float32)

# --- FUNÇÃO PARA EMBEDDINGS BERT (sentence-transformers genérico) ---
def embed_bert(texts, model):
    return np.array(model.encode(texts, convert_to_numpy=True), dtype=np.float32)

# --- FUNÇÃO PARA EMBEDDINGS LEGALBERT PT-BR ---
class LegalBERTEmbedder:
    def __init__(self, model_name="dominguesm/legal-bert-base-cased-ptbr"):
        self.tokenizer = AutoTokenizer.from_pretrained(model_name)
        self.model = AutoModel.from_pretrained(model_name)
        self.model.eval()
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.model.to(self.device)

    def embed(self, texts):
        embeddings = []
        with torch.no_grad():
            for t in texts:
                inputs = self.tokenizer(t, truncation=True, padding=True, return_tensors="pt", max_length=512)
                inputs = {k: v.to(self.device) for k,v in inputs.items()}
                outputs = self.model(**inputs)
                cls_emb = outputs.last_hidden_state[:,0,:].squeeze().cpu().numpy()
                embeddings.append(cls_emb)
        return np.array(embeddings, dtype=np.float32)

# --- FUNÇÃO PARA EMBEDDINGS WORD2VEC ---
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

# --- FUNÇÃO PARA CRIAR ÍNDICE FAISS ---
def build_index(embeddings):
    dim = embeddings.shape[1]
    index = faiss.IndexFlatIP(dim)
    faiss.normalize_L2(embeddings)
    index.add(embeddings)
    return index

# --- FUNÇÃO DE BUSCA ---
def search(index, query_emb, n_docs):
    start = time.time()
    faiss.normalize_L2(query_emb)
    scores, idx = index.search(query_emb, n_docs)
    elapsed = int((time.time() - start) * 1_000_000)  # microssegundos
    return idx[0].astype(np.uint16), elapsed

# --- CARREGAR INPUTS ---
def load_inputs(json_path):
    with open(json_path, encoding="utf-8") as f:
        return json.load(f)

# --- MAIN EXECUTION ---
# /Users/arthurandrade/Desktop/SENAC/Tcc2/src/misc/corpus/clean
laws_folder = "../misc/corpus/clean"
inputs_path = "../misc/searchLegalInputs.json"
files, texts = load_texts(laws_folder)
n_docs = len(files)
TOP_K = 20

print(f"{n_docs} arquivos carregados.")

# Gerar embeddings para os documentos
print("Gerando embeddings dos documentos...")

# OpenAI
# embeddings_openai = embed_openai(texts)
# index_openai = build_index(embeddings_openai)

# BERT genérico
bert_model = SentenceTransformer('all-MiniLM-L6-v2')
embeddings_bert = embed_bert(texts, bert_model)
index_bert = build_index(embeddings_bert)

# LegalBERT
# legalbert_model = LegalBERTEmbedder()
# embeddings_legalbert = legalbert_model.embed(texts)
# index_legalbert = build_index(embeddings_legalbert)

# Word2Vec
w2v_model = Word2VecEmbedder()
embeddings_w2v = w2v_model.embed(texts)
index_w2v = build_index(embeddings_w2v)

# Carregar inputs
inputs = load_inputs(inputs_path)
final_output = {"words10": [], "words20": [], "words40": []}

# Função para gerar sample
def generate_sample(phrase):
    sample = {"input": phrase}

    # Embeddings da frase
    # query_openai = embed_openai([phrase])
    query_bert = embed_bert([phrase], bert_model)
    # query_legalbert = legalbert_model.embed([phrase])
    query_w2v = w2v_model.embed([phrase])

    # Busca e tempos
    # ordered, t = search(index_openai, query_openai, n_docs)
    # sample["openai"] = ordered.tolist()
    # sample["openaiT"] = t

    ordered, t = search(index_bert, query_bert, n_docs)
    sample["bert"] = ordered[:TOP_K].tolist()
    sample["bertT"] = t

    # ordered, t = search(index_legalbert, query_legalbert, n_docs)
    # sample["legalBert"] = ordered[:TOP_K].tolist()
    # sample["legalBertT"] = t

    ordered, t = search(index_w2v, query_w2v, n_docs)
    sample["word2vec"] = ordered[:TOP_K].tolist()
    sample["word2vecT"] = t

    return sample

# Gerar JSON final
for group in ["words10", "words20", "words40"]:
    print(f"Processando grupo {group}...")
    for phrase in tqdm(inputs[group]):
        final_output[group].append(generate_sample(phrase))

# Salvar JSON
with open("output.json", "w", encoding="utf-8") as f:
    json.dump(final_output, f, ensure_ascii=False, indent=2)

print("Finalizado! JSON salvo como output.json")
