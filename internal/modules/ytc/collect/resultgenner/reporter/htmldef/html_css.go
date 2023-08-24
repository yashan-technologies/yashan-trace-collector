package htmldef

const HTML_CSS = `
<style type="text/css">
    * {
        margin: 0;
        padding: 0;
    }

    #catalogs {
        display: none;
    }

    body {
        padding: 20px 80px;
        background-color: #f5f7f9;
        color: #000;
        font-family: Verdana
    }

    table {
        border-collapse: collapse;
        border-spacing: 0;
        width: 100%;
        margin-bottom: 20px;
    }

    .ytc_table {
        border: solid #ccc 1px;
        border-radius: 10px;
        box-shadow: 0 1px 1px #ccc;
        overflow: hidden;
    }

    .ytc_table caption {
        font-size: larger;
        font-weight: bold;
    }

    .ytc_table tr {
        transition: all 0.1s ease-in-out;
    }

    .ytc_table tr:nth-child(even) {
        background-color: #f9f9f9;
    }

    .ytc_table tr:nth-child(odd) {
        background-color: #ffffff;
    }

    .ytc_table .highlight,
    .ytc_table tr:hover {
        background: #f0ddbe;
    }

    .ytc_table td,
    .ytc_table th {
        border-left: 1px solid #ccc;
        border-top: 1px solid #ccc;
        padding: 10px;
        text-align: left;
    }

    .ytc_table th {
        background-color: #f5b31a;
        background-image: linear-gradient(top, #ffe194, #f5b31a);
        box-shadow: 0 1px 0 rgba(255, 255, 255, .8) inset;
        border-top: none;
        text-shadow: 0 1px 0 rgba(255, 255, 255, .5);
    }

    .ytc_table td:first-child,
    .ytc_table th:first-child {
        border-left: none;
    }

    .ytc_table th:first-child {
        border-radius: 10px 0 0 0;
    }

    .ytc_table th:last-child {
        border-radius: 0 10px 0 0;
    }

    .ytc_table tr:last-child td:first-child {
        border-radius: 0 0 0 10px;
    }

    .ytc_table tr:last-child td:last-child {
        border-radius: 0 0 10px 0;
    }


    /* ytc_list */
    ol,
    ul {
        width: 100%;
        display: inline-block;
        text-align: left;
        vertical-align: top;
        background: #f5f7f9;
        border-radius: 5px;
        padding: 0.5em;
    }

    li {
        list-style: none;
        position: relative;
        padding: 0 0 0 1em;
        margin: 0 0 .2em 0;
        -webkit-transition: .12s;
        transition: .12s;
    }

    li::before {
        position: absolute;
        content: '\2022';
        top: 0.25em;
        left: 0;
        text-align: center;
        font-size: 1em;
        opacity: .5;
        line-height: .75;
        -webkit-transition: .5s;
        transition: .5s;
    }

    li:hover {
        color: #f5b31a;
        font-size: 1.1em;
    }

    li:hover::before {
        -webkit-transform: scale(1.1);
        -ms-transform: scale(1.1);
        transform: scale(1.1);
        opacity: 1;
        text-shadow: 0 0 4px;
        -webkit-transition: .1s;
        transition: .1s;
    }

    .ytc_list>li {
        font-weight: bold;
        font-size: 1.2em;
        padding: 0;
    }

    .ytc_list>li::before {
        content: "";
    }

    /* ytc_button */
    .ytc_button {
        background-color: #F5B31A;
        color: #000000;
        padding: 10px 20px;
        border: none;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        cursor: pointer;
        transition: background-color 0.3s ease;
        border-radius: 15px;
        font-family: Verdana, sans-serif;
        font-weight: bold;
        margin-bottom: 10px;
    }

    .ytc_button:hover {
        background-color: #ffc94a;
    }

    .ytc_button:active {
        transform: translateY(1px);
    }
</style>
`
