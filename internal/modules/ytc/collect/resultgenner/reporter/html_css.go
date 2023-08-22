package reporter

const HTML_CSS = `
<style type="text/css">
    * {
        margin: 0;
        padding: 0;
    }

    body {
        padding: 20px 80px;
        background-color: #f5f7f9;
        color: #000;
    }

    table {
        *border-collapse: collapse;
        /* IE7 and lower */
        border-spacing: 0;
        width: 100%;
        margin-bottom: 20px;
    }

    /* ytc_table */
    .ytc_table {
        border: solid #ccc 1px;
        -moz-border-radius: 6px;
        -webkit-border-radius: 6px;
        border-radius: 6px;
        -webkit-box-shadow: 0 1px 1px #ccc;
        -moz-box-shadow: 0 1px 1px #ccc;
        box-shadow: 0 1px 1px #ccc;
    }

    .ytc_table caption {
        font-size: larger;
        font-style: initial;
        font-weight: bold;
    }

    .ytc_table tr {
        -o-transition: all 0.1s ease-in-out;
        -webkit-transition: all 0.1s ease-in-out;
        -moz-transition: all 0.1s ease-in-out;
        -ms-transition: all 0.1s ease-in-out;
        transition: all 0.1s ease-in-out;
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
        background-image: -webkit-gradient(linear, left top, left bottom, from(#ffe194), to(#f5b31a));
        background-image: -webkit-linear-gradient(top, #ffe194, #f5b31a);
        background-image: -moz-linear-gradient(top, #ffe194, #f5b31a);
        background-image: -ms-linear-gradient(top, #ffe194, #f5b31a);
        background-image: -o-linear-gradient(top, #ffe194, #f5b31a);
        background-image: linear-gradient(top, #ffe194, #f5b31a);
        filter: progid:DXImageTransform.Microsoft.gradient(GradientType=0, startColorstr=#ffe194, endColorstr=#f5b31a);
        -ms-filter: "progid:DXImageTransform.Microsoft.gradient (GradientType=0, startColorstr=#ffe194, endColorstr=#f5b31a)";
        -webkit-box-shadow: 0 1px 0 rgba(255, 255, 255, .8) inset;
        -moz-box-shadow: 0 1px 0 rgba(255, 255, 255, .8) inset;
        box-shadow: 0 1px 0 rgba(255, 255, 255, .8) inset;
        border-top: none;
        text-shadow: 0 1px 0 rgba(255, 255, 255, .5);
    }

    .ytc_table td:first-child,
    .ytc_table th:first-child {
        border-left: none;
    }

    .ytc_table th:first-child {
        -moz-border-radius: 6px 0 0 0;
        -webkit-border-radius: 6px 0 0 0;
        border-radius: 6px 0 0 0;
    }

    .ytc_table th:last-child {
        -moz-border-radius: 0 6px 0 0;
        -webkit-border-radius: 0 6px 0 0;
        border-radius: 0 6px 0 0;
    }

    .ytc_table tr:last-child td:first-child {
        -moz-border-radius: 0 0 0 6px;
        -webkit-border-radius: 0 0 0 6px;
        border-radius: 0 0 0 6px;
    }

    .ytc_table tr:last-child td:last-child {
        -moz-border-radius: 0 0 6px 0;
        -webkit-border-radius: 0 0 6px 0;
        border-radius: 0 0 6px 0;
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
        font-family: Arial;
        top: 0;
        left: 0;
        text-align: center;
        font-size: 2em;
        opacity: .5;
        line-height: .75;
        -webkit-transition: .5s;
        transition: .5s;
    }

    li:hover {
        color: #f5b31a;
        font-size: 1.5em;
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
</style>
`
