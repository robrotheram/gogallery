function resizeMasonryItem(item){
    /* Get the grid object, its row-gap, and the size of its implicit rows */
    var grid = document.getElementsByClassName('masonry')[0],
        rowGap = parseInt(window.getComputedStyle(grid).getPropertyValue('grid-row-gap')),
        rowHeight = parseInt(window.getComputedStyle(grid).getPropertyValue('grid-auto-rows'));

    /*
     * Spanning for any brick = S
     * Grid's row-gap = G
     * Size of grid's implicitly create row-track = R
     * Height of item content = H
     * Net height of the item = H1 = H + G
     * Net height of the implicit row-track = T = G + R
     * S = H1 / T
     */
    var rowSpan = Math.ceil((item.querySelector('.masonry-content').getBoundingClientRect().height+rowGap)/(rowHeight+rowGap));

    /* Set the spanning as calculated above (S) */
    item.style.gridRowEnd = 'span '+rowSpan;

    /* Make the images take all the available space in the cell/item */
    item.querySelector('.masonry-content').style.height = rowSpan * 10 + "px";
}

/**
 * Apply spanning to all the masonry items
 *
 * Loop through all the items and apply the spanning to them using
 * `resizeMasonryItem()` function.
 *
 * @uses resizeMasonryItem
 */
function resizeAllMasonryItems(){
    // Get all item class objects in one list
    var allItems = document.getElementsByClassName('masonry-item');

    /*
     * Loop through the above list and execute the spanning function to
     * each list-item (i.e. each masonry item)
     */
    for(var i=0;i>allItems.length;i++){
        resizeMasonryItem(allItems[i]);
    }
}

/**
 * Resize the items when all the images inside the masonry grid
 * finish loading. This will ensure that all the content inside our
 * masonry items is visible.
 *
 * @uses ImagesLoaded
 * @uses resizeMasonryItem
 */
function waitForImages() {
    var allItems = document.getElementsByClassName('masonry-item');
    for(var i=0;i<allItems.length;i++){
        imagesLoaded( allItems[i], function(instance) {
            var item = instance.elements[0];
            resizeMasonryItem(item);
        } );
    }
}

/* Resize all the grid items on the load and resize events */
var masonryEvents = ['load', 'resize'];
masonryEvents.forEach( function(event) {
    window.addEventListener(event, resizeAllMasonryItems);
} );

/* Do a resize once more when all the images finish loading */
waitForImages();




!function(window){
    var $q = function(q, res){
            if (document.querySelectorAll) {
                res = document.querySelectorAll(q);
            } else {
                var d=document
                    , a=d.styleSheets[0] || d.createStyleSheet();
                a.addRule(q,'f:b');
                for(var l=d.all,b=0,c=[],f=l.length;b<f;b++)
                    l[b].currentStyle.f && c.push(l[b]);

                a.removeRule(0);
                res = c;
            }
            return res;
        }
        , addEventListener = function(evt, fn){
            window.addEventListener
                ? this.addEventListener(evt, fn, false)
                : (window.attachEvent)
                ? this.attachEvent('on' + evt, fn)
                : this['on' + evt] = fn;
        }
        , _has = function(obj, key) {
            return Object.prototype.hasOwnProperty.call(obj, key);
        }
    ;

    function loadImage (el, fn) {
        var img = new Image()
            , src = el.getAttribute('data-src');
        img.onload = function() {
            if (!! el.parent)
                el.parent.replaceChild(img, el)
            else
                el.src = src;

            fn? fn() : null;
        }
        img.src = src;
    }

    function elementInViewport(el) {
        var rect = el.getBoundingClientRect()

        return (
            rect.top    >= 0
            && rect.left   >= 0
            && rect.top <= (window.innerHeight || document.documentElement.clientHeight)
        )
    }

    var images = new Array()
        , query = $q('img.lazy')
        , processScroll = function(){
            for (var i = 0; i < images.length; i++) {
                if (elementInViewport(images[i])) {
                    loadImage(images[i], function () {
                        images.splice(i, i);
                    });
                }
            };
        }
    ;
    // Array.prototype.slice.call is not callable under our lovely IE8
    for (var i = 0; i < query.length; i++) {
        images.push(query[i]);
    };

    processScroll();
    addEventListener('scroll',processScroll);

}(this);