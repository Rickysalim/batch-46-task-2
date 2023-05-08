function onSubmit() {
  const name = document.getElementById("name").value;
  const email = document.getElementById("email").value;
  const phone = document.getElementById("phone").value;
  const subject = document.getElementById("subject").value;
  const message = document.getElementById("message").value;

  // validation
  if (!name) {
    alert("Nama Harus Di Isi");
  } else if (!email) {
    alert("Email Harus Di Isi");
  } else if (!phone) {
    alert("Nomor Telepon Harus Di Isi");
  } else if (!subject) {
    alert("Subject Harus Di Isi");
  } else if (!message) {
    alert("Message Harus Di Isi");
  }

  const destination = "rickysalim81@gmail.com";
  let a = document.createElement("a");
  a.href = `mailto:${destination}?subject=${subject}&body= Hallo nama saya ${name}, saya ingin ${message}, bisakah anda menghubungi saya di ${phone}`;
  a.click();
}
