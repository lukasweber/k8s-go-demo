import './App.css';
import { useState } from "react";

function App() {

  const [from, setFrom] = useState(0);
  const [to, setTo] = useState(0);
  const [results, setResults] = useState([]);

  function handleCalculate() {
    const index = results.length;
    setResults(prevResults => [...prevResults, { isLoading: true, key: index }])
    fetch(`/api/prime?from=${from}&to=${to}`)
      .then(res => res.json())
      .then(
        result => {
          setResults((items) => {
            const newItems = [...items];
            newItems[index] = {
              ...newItems[index],
              isLoading: false,
              apiPod: result["ApiHostName"].split("-").pop(),
              calcPod: result["RpcHostName"].split("-").pop(),
              count: result["Count"]
            }
            return newItems;
          });
        }
      )
  }

  return (
    <div className="App">
      <section className="App-container">
        <div className="App-form">
          <div className="form-group">
            <label htmlFor="from">From</label>
            <input type="number" name="from" value={from} onChange={e => setFrom(e.target.value)}></input>
          </div>
          <div className="form-group">
            <label htmlFor="to">To</label>
            <input type="number" name="to" value={to} onChange={e => setTo(e.target.value)}></input>
          </div>
          <div className="form-group">
            <button onClick={handleCalculate}>Calculate</button>
          </div>
        </div>
        <div className="App-results">
          {results.map(item => item && item.isLoading ?
            (
              <div className="App-results-item" key={item.key}>Loading ...</div>
            ) :
            (
              <div className="App-results-item" key={item.key}>
                <span>Result: <b>{item.count}</b></span>
                <span>(Api Pod: {item.apiPod}, Calc Pod: {item.calcPod})</span>
              </div>
            )
          )}
        </div>
      </section>
    </div>
  );
}

export default App;
