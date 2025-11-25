async function loadPosts() {
    const res = await fetch("http://localhost:8080/api/v1/posts");
    const data = await res.json();
    document.getElementById("output").textContent = JSON.stringify(data, null, 2);
}
