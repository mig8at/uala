const fs = require('fs');
const path = require('path');

// Función para generar el símbolo adecuado dependiendo de si es el último elemento
function obtenerSimbolo(indice, total) {
    return indice === total - 1 ? '└── ' : '├── ';
}

// Función asíncrona para leer y procesar directorios de forma recursiva
async function leerDirectorioRecursivamente(directorio, prefijo = '') {
    try {
        const items = await fs.promises.readdir(directorio, { withFileTypes: true });
        const total = items.length;

        for (let i = 0; i < total; i++) {
            const item = items[i];
            const rutaCompleta = path.join(directorio, item.name);
            const esUltimo = i === total - 1;
            const simbolo = obtenerSimbolo(i, total);
            const nuevaPrefijo = prefijo + (esUltimo ? '    ' : '│   ');

            if (item.isDirectory()) {
                await leerDirectorioRecursivamente(rutaCompleta, nuevaPrefijo);
            } else if (item.isFile()) {
                if (item.name.endsWith('.freezed.dart') || item.name.endsWith('.g.dart')) {
                    // Mostrar '...' para los archivos excluidos
                    console.log(`${prefijo}${simbolo}${item.name} -> ...`);
                } else {
                    try {
                        let contenido = await fs.promises.readFile(rutaCompleta, 'utf8');
                        contenido = contenido.replace(/[\r\n]+/g, ' '); // Eliminar saltos de línea
                        console.log(`${prefijo}${simbolo}${item.name} -> ${contenido}`);
                    } catch (error) {
                        console.error(`${prefijo}${simbolo}Error al leer el archivo ${item.name}:`, error.message);
                    }
                }
            }
        }
    } catch (error) {
        console.error(`Error al leer el directorio ${directorio}:`, error.message);
    }
}

// Función para escribir la estructura en un archivo de texto
async function escribirEstructuraEnArchivo(directorio, archivoSalida) {
    const stream = fs.createWriteStream(archivoSalida, { encoding: 'utf8' });

    // Función interna para manejar la escritura en el stream
    async function escribirDirectorio(directorio, prefijo = '') {
        try {
            const items = await fs.promises.readdir(directorio, { withFileTypes: true });
            const total = items.length;

            for (let i = 0; i < total; i++) {
                const item = items[i];
                const rutaCompleta = path.join(directorio, item.name);
                const esUltimo = i === total - 1;
                const simbolo = obtenerSimbolo(i, total);
                const nuevaPrefijo = prefijo + (esUltimo ? '    ' : '│   ');

                if (item.isDirectory()) {
                    stream.write(`${prefijo}${simbolo}${item.name}/\n`);
                    await escribirDirectorio(rutaCompleta, nuevaPrefijo);
                } else if (item.isFile()) {
                    if (item.name.endsWith('.freezed.dart') || item.name.endsWith('_mock.go')) {
                        // Escribir '...' para los archivos excluidos
                        stream.write(`${prefijo}${simbolo}${item.name} -> ...\n`);
                    } else {
                        try {
                            let contenido = await fs.promises.readFile(rutaCompleta, 'utf8');
                            contenido = contenido.replace(/[\r\n]+/g, ' '); // Eliminar saltos de línea
                            stream.write(`${prefijo}${simbolo}${item.name} -> ${contenido}\n`);
                        } catch (error) {
                            stream.write(`${prefijo}${simbolo}Error al leer el archivo ${item.name}: ${error.message}\n`);
                        }
                    }
                }
            }
        } catch (error) {
            stream.write(`Error al leer el directorio ${directorio}: ${error.message}\n`);
        }
    }

    await escribirDirectorio(directorio);
    stream.end();
    console.log(`La estructura se ha guardado en ${archivoSalida}`);
}

// Ruta al directorio 'src' (asegúrate de que exista)
const rutaSrc = path.join(__dirname, 'internal');

// Ejecutar y mostrar en consola
console.log('Estructura de Directorios y Archivos:\n');
leerDirectorioRecursivamente(rutaSrc).then(() => {
    console.log('\nEstructura completada en la consola.');
});

// Ejecutar y guardar en un archivo de texto
const archivoSalida = 'estructura.txt';
escribirEstructuraEnArchivo(rutaSrc, archivoSalida);
