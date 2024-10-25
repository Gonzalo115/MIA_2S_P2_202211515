import React, { useState, useRef } from "react"

export default function Component() {
  const [inputText, setInputText] = useState("");
  const [outputText, setOutputText] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setInputText(content);
      };
      reader.readAsText(file);
    }
  };

  const handleExecute = async () => {
    try {
      const response = await fetch("http://localhost:3000/analyze", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command: inputText }),
      });

      if (!response.ok) {
        throw new Error("Network response was not ok");
      }

      const data = await response.json();
      const results = data.results.join("\n");
      setOutputText(results);
    } catch (error) {
      console.error("Error:", error);
      setOutputText(`Error: ${error}`);
    }
  };
  const handleClearInput = () => {
    setInputText("");
  };


  return (
    <div className="flex flex-col min-h-screen bg-gray-600">
      <div className="flex-grow flex items-center justify-center p-4">
        <div className="w-full max-w-3xl p-8 bg-white rounded-lg shadow-md">
          {/* Título centrado */}
          <h1 className="text-3xl font-semibold text-center mb-6 text-gray-800 uppercase font-sans">
            GESTOR DE DISCOS
          </h1>
    
          {/* Botón Examinar a la izquierda con espacio abajo */}
          <div className="flex justify-start mb-4">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              className="hidden"
              accept=".smia"
            />
            <button
              onClick={() => fileInputRef.current?.click()}
              className="px-4 py-2 bg-white text-black text-lg border-4 border-black rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2"
            >
              Examinar
            </button>

            <button
              onClick={handleClearInput}
              className="px-4 py-2 bg-white text-black text-lg border-3 border-black rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 ml-auto"
            >
              🗑️
            </button>
          </div>
    
          {/* Textarea para el terminal de entrada */}
          <div className="mb-4">
            <textarea
              className="w-full h-48 p-2 border-4 border-black rounded-md resize-none bg-gray-200 focus:outline-none focus:border-gray-800"
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              placeholder="Terminal de entrada"
            />
          </div>
    
          {/* Textarea para el terminal de salida */}
          <div className="mb-4">
            <textarea
              className="w-full h-64 p-2 border border-gray-300 rounded-md resize-none bg-black text-white font-mono text-xs focus:outline-none"
              value={outputText}
              readOnly
              placeholder="Terminal de salida"
            />
          </div>
    
          {/* Botón Ejecutar */}
          <div className="flex justify-between">
            <div>
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileChange}
                className="hidden"
                accept=".txt"
              />
            </div>
            <button
              onClick={handleExecute}
              className="w-full px-4 py-2 bg-white text-black text-lg border-4 border-black rounded-md hover:bg-gray-200 focus:outline-none"
            >
              Ejecutar
            </button>
          </div>
        </div>
      </div>
    </div>
  );
  
  
}
